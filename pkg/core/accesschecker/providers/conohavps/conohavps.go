package conohavps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/accesschecker"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
)

const (
	defaultV2IdentityURL = "https://identity.tyo1.conoha.io/v2.0/tokens"
	defaultV3IdentityURL = "https://identity.c3j1.conoha.io/v3/auth/tokens"
)

type checker struct {
	v2IdentityURL string
	v3IdentityURL string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{
		v2IdentityURL: defaultV2IdentityURL,
		v3IdentityURL: defaultV3IdentityURL,
	}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeConoHaVPS)
}

func (c *checker) Scope() []string {
	return []string{core.AccessCheckItemConnect, core.AccessCheckItemAuth}
}

type checkConfig struct {
	ApiVersion  string
	ApiUserId   string
	ApiUserName string
	ApiPassword string
	TenantId    string
	TenantName  string
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForConoHaVPS
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		ApiVersion:  normalizeVersion(wire.ApiVersion),
		ApiUserId:   strings.TrimSpace(wire.ApiUserId),
		ApiUserName: strings.TrimSpace(wire.ApiUserName),
		ApiPassword: wire.ApiPassword,
		TenantId:    strings.TrimSpace(wire.TenantId),
		TenantName:  strings.TrimSpace(wire.TenantName),
	}
	switch cfg.ApiVersion {
	case "v2":
		if cfg.ApiUserName == "" {
			return nil, fmt.Errorf("%w: apiUserName is required for api version v2", core.ErrAccessCheckConfigInvalid)
		}
		if cfg.TenantId == "" {
			return nil, fmt.Errorf("%w: tenantId is required for api version v2", core.ErrAccessCheckConfigInvalid)
		}
	case "v3":
		if cfg.ApiUserId == "" && cfg.ApiUserName == "" {
			return nil, fmt.Errorf("%w: apiUserId or apiUserName is required for api version v3", core.ErrAccessCheckConfigInvalid)
		}
		if cfg.TenantId == "" && cfg.TenantName == "" {
			return nil, fmt.Errorf("%w: tenantId or tenantName is required for api version v3", core.ErrAccessCheckConfigInvalid)
		}
	default:
		return nil, fmt.Errorf("%w: unsupported api version '%s'", core.ErrAccessCheckConfigInvalid, wire.ApiVersion)
	}
	if strings.TrimSpace(cfg.ApiPassword) == "" {
		return nil, fmt.Errorf("%w: apiPassword is required", core.ErrAccessCheckConfigInvalid)
	}
	return cfg, nil
}

func normalizeVersion(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimSuffix(v, ".0")
	if v != "" {
		v = "v" + v
	}
	return v
}

func (c *checker) Check(ctx context.Context, config map[string]any) (*CheckResult, error) {
	ctx, cancel := accesschecker.WithBudget(ctx)
	defer cancel()

	cfg, err := parseConfig(config)
	if err != nil {
		return nil, err
	}

	var endpoint string
	var payload any
	if cfg.ApiVersion == "v2" {
		endpoint = c.v2IdentityURL
		payload = map[string]any{
			"auth": map[string]any{
				"tenantId": cfg.TenantId,
				"passwordCredentials": map[string]any{
					"username": cfg.ApiUserName,
					"password": cfg.ApiPassword,
				},
			},
		}
	} else {
		endpoint = c.v3IdentityURL
		user := map[string]any{"password": cfg.ApiPassword}
		if cfg.ApiUserId != "" {
			user["id"] = cfg.ApiUserId
		}
		if cfg.ApiUserName != "" {
			user["name"] = cfg.ApiUserName
		}
		auth := map[string]any{
			"identity": map[string]any{
				"methods": []string{"password"},
				"password": map[string]any{
					"user": user,
				},
			},
		}
		if cfg.TenantId != "" || cfg.TenantName != "" {
			project := map[string]any{}
			if cfg.TenantId != "" {
				project["id"] = cfg.TenantId
			}
			if cfg.TenantName != "" {
				project["name"] = cfg.TenantName
			}
			auth["scope"] = map[string]any{"project": project}
		}
		payload = map[string]any{"auth": auth}
	}

	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, fmt.Errorf("%w: identity endpoint is invalid", core.ErrAccessCheckConfigInvalid)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%w: request body is not serializable", core.ErrAccessCheckConfigInvalid)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: identity endpoint is invalid", core.ErrAccessCheckConfigInvalid)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	diagnostics := []string{u.Host}
	secrets := []string{cfg.ApiPassword, cfg.ApiUserId, cfg.ApiUserName, cfg.TenantId, cfg.TenantName}
	client := accesschecker.HTTPClient()
	client.Timeout = accesschecker.BudgetLeft(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return accesschecker.Finish(err, 0, "", diagnostics, secrets...), nil
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))

	diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", resp.StatusCode))
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		diag = append(diag, "identity token issued; token value discarded")
	}
	return accesschecker.Finish(nil, resp.StatusCode, "", diag, secrets...), nil
}
