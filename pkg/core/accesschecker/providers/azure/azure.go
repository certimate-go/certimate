package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	_ "github.com/Azure/azure-sdk-for-go/sdk/azcore/arm/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/accesschecker"
	xazure "github.com/certimate-go/certimate/pkg/utils/third-party/azure"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
)

const armAPIVersion = "2022-12-01"

type checker struct {
	authority string
	arm       string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeAzure)
}

func (c *checker) Scope() []string {
	return []string{core.AccessCheckItemConnect, core.AccessCheckItemAuth}
}

func (c *checker) Check(ctx context.Context, config map[string]any) (*CheckResult, error) {
	ctx, cancel := accesschecker.WithBudget(ctx)
	defer cancel()

	cfg, err := parseConfig(config)
	if err != nil {
		return nil, err
	}

	authority, arm, scope, err := resolveEndpoints(cfg.CloudName, c.authority, c.arm)
	if err != nil {
		return nil, err
	}

	diagnostics := []string{accesschecker.HostOf(authority), accesschecker.HostOf(arm)}
	secrets := []string{cfg.ClientId, cfg.ClientSecret}
	token, status, code, callErr := acquireToken(ctx, authority, cfg.TenantId, cfg.ClientId, cfg.ClientSecret, scope)
	if callErr != nil || token == "" {
		return accesschecker.Finish(callErr, status, code, diagnostics, secrets...), nil
	}

	status, callErr = listSubscriptions(ctx, arm, cfg.SubscriptionId, token)
	secrets = append(secrets, token)
	if callErr != nil {
		return accesschecker.Finish(callErr, status, "", diagnostics, secrets...), nil
	}
	if status == http.StatusForbidden {
		diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", status))
		return accesschecker.Redact(accesschecker.Passed(diag), secrets...), nil
	}
	return accesschecker.Finish(nil, status, "", diagnostics, secrets...), nil
}

type checkConfig struct {
	TenantId       string
	ClientId       string
	ClientSecret   string
	SubscriptionId string
	CloudName      string
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForAzure
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		TenantId:       strings.TrimSpace(wire.TenantId),
		ClientId:       strings.TrimSpace(wire.ClientId),
		ClientSecret:   wire.ClientSecret,
		SubscriptionId: strings.TrimSpace(wire.SubscriptionId),
		CloudName:      strings.TrimSpace(wire.CloudName),
	}
	if cfg.TenantId == "" || !safePathSegment(cfg.TenantId) {
		return nil, fmt.Errorf("%w: tenantId is required", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.ClientId == "" {
		return nil, fmt.Errorf("%w: clientId is required", core.ErrAccessCheckConfigInvalid)
	}
	if strings.TrimSpace(cfg.ClientSecret) == "" {
		return nil, fmt.Errorf("%w: clientSecret is required", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.SubscriptionId != "" && !safePathSegment(cfg.SubscriptionId) {
		return nil, fmt.Errorf("%w: subscriptionId is invalid", core.ErrAccessCheckConfigInvalid)
	}
	if _, err := xazure.GetCloudEnvConfiguration(cfg.CloudName); err != nil {
		return nil, fmt.Errorf("%w: cloudName is invalid", core.ErrAccessCheckConfigInvalid)
	}
	return cfg, nil
}

func resolveEndpoints(cloudName, authorityOverride, armOverride string) (string, string, string, error) {
	env, err := xazure.GetCloudEnvConfiguration(cloudName)
	if err != nil {
		return "", "", "", fmt.Errorf("%w: cloudName is invalid", core.ErrAccessCheckConfigInvalid)
	}
	svc := env.Services[cloud.ResourceManager]
	if svc.Endpoint == "" || svc.Audience == "" || env.ActiveDirectoryAuthorityHost == "" {
		return "", "", "", fmt.Errorf("%w: cloudName is invalid", core.ErrAccessCheckConfigInvalid)
	}
	authority := strings.TrimRight(env.ActiveDirectoryAuthorityHost, "/")
	arm := strings.TrimRight(svc.Endpoint, "/")
	if authorityOverride != "" {
		authority = strings.TrimRight(authorityOverride, "/")
	}
	if armOverride != "" {
		arm = strings.TrimRight(armOverride, "/")
	}
	scope := strings.TrimRight(svc.Audience, "/") + "/.default"
	return authority, arm, scope, nil
}

func acquireToken(ctx context.Context, authority, tenantId, clientId, clientSecret, scope string) (string, int, string, error) {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", clientId)
	form.Set("client_secret", clientSecret)
	form.Set("scope", scope)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authority+"/"+url.PathEscape(tenantId)+"/oauth2/v2.0/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", 0, "", fmt.Errorf("%w: token url is invalid", core.ErrAccessCheckConfigInvalid)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := accesschecker.HTTPClient().Do(req)
	if err != nil {
		return "", 0, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", resp.StatusCode, "", err
	}
	var payload struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	_ = json.Unmarshal(body, &payload)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && payload.AccessToken != "" {
		return payload.AccessToken, resp.StatusCode, "", nil
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return "", 0, "", fmt.Errorf("token endpoint returned no credential verdict")
	}
	return "", resp.StatusCode, payload.Error, fmt.Errorf("token endpoint rejected the credential")
}

func listSubscriptions(ctx context.Context, arm, subscriptionId, token string) (int, error) {
	target := arm + "/subscriptions?api-version=" + armAPIVersion
	if subscriptionId != "" {
		target = arm + "/subscriptions/" + url.PathEscape(subscriptionId) + "?api-version=" + armAPIVersion
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return 0, fmt.Errorf("%w: arm url is invalid", core.ErrAccessCheckConfigInvalid)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := accesschecker.HTTPClient().Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	return resp.StatusCode, nil
}

func safePathSegment(value string) bool {
	if value == "" || strings.Contains(value, "..") {
		return false
	}
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '.' || r == '_' || r == '-' || r == ':':
		default:
			return false
		}
	}
	return true
}
