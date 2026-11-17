package rucenter

import (
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

const apiBaseURL = "https://api.nic.ru"

type checker struct {
	baseURL string
}

type checkConfig struct {
	Username string
	Password string
	ApplicationId string
	ApplicationToken string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{baseURL: apiBaseURL}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeRuCenter)
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

	req, err := c.newProbe(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: probe url is invalid", core.ErrAccessCheckConfigInvalid)
	}

	diagnostics := []string{req.URL.Host}
	client := accesschecker.HTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return accesschecker.Finish(err, 0, "", diagnostics, cfg.redact()...), nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", resp.StatusCode), "response body unreadable")
		return accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, diag), cfg.redact()...), nil
	}

	if resp.StatusCode == http.StatusOK {
		var tok struct {
			AccessToken string `json:"access_token"`
			Error       string `json:"error"`
		}
		if err := json.Unmarshal(body, &tok); err != nil || tok.AccessToken == "" {
			if tok.Error != "" {
				return c.classify(resp.StatusCode, body, diagnostics, cfg), nil
			}
			return remoteResult(diagnostics, resp.StatusCode, cfg.redact(), "no credential verdict"), nil
		}
		return c.services(ctx, client, cfg, tok.AccessToken, diagnostics)
	}
	return c.classify(resp.StatusCode, body, diagnostics, cfg), nil
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForRuCenter
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		Username: strings.TrimSpace(wire.Username),
		Password: strings.TrimSpace(wire.Password),
		ApplicationId: strings.TrimSpace(wire.ApplicationId),
		ApplicationToken: strings.TrimSpace(wire.ApplicationToken),
	}
	if cfg.Username == "" {
		return nil, fmt.Errorf("%w: username is required", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.Password == "" {
		return nil, fmt.Errorf("%w: password is required", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.ApplicationId == "" {
		return nil, fmt.Errorf("%w: applicationId is required", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.ApplicationToken == "" {
		return nil, fmt.Errorf("%w: applicationToken is required", core.ErrAccessCheckConfigInvalid)
	}

	return cfg, nil
}

func (c *checkConfig) redact() []string {
	return []string{c.Username, c.Password, c.ApplicationId, c.ApplicationToken}
}

func (c *checker) newProbe(ctx context.Context, cfg *checkConfig) (*http.Request, error) {
u, err := c.endpoint("oauth", "token")
if err != nil {
	return nil, err
}
form := url.Values{}
form.Set("grant_type", "password")
form.Set("username", cfg.Username)
form.Set("password", cfg.Password)
form.Set("client_id", cfg.ApplicationId)
form.Set("client_secret", cfg.ApplicationToken)
form.Set("scope", ".+:/dns-master/.+")
req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), strings.NewReader(form.Encode()))
if err != nil {
	return nil, err
}
req.Header.Set("Accept", "application/json")
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
return req, nil
}

func (c *checker) endpoint(elem ...string) (*url.URL, error) {
	base := c.baseURL
	if strings.TrimSpace(base) == "" {
		base = apiBaseURL
	}
	u, err := url.Parse(base)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, fmt.Errorf("probe url is invalid")
	}
	if len(elem) > 0 {
		u = u.JoinPath(elem...)
	}
	return u, nil
}

func gateStatus(status int, diagnostics []string, secrets []string, serverErrors bool) *CheckResult {
	if status == http.StatusUnauthorized || status == http.StatusForbidden || (serverErrors && status >= 500) || (status >= 300 && status < 400) {
		return accesschecker.Finish(nil, status, "", diagnostics, secrets...)
	}
	return nil
}

func withStatus(diagnostics []string, status int, extra ...string) []string {
	out := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", status))
	return append(out, extra...)
}

func passResult(diagnostics []string, status int, secrets []string, extra ...string) *CheckResult {
	return accesschecker.Redact(accesschecker.Passed(withStatus(diagnostics, status, extra...)), secrets...)
}

func authResult(diagnostics []string, status int, secrets []string, extra ...string) *CheckResult {
	return accesschecker.Redact(accesschecker.AuthRejected(withStatus(diagnostics, status, extra...)), secrets...)
}

func remoteResult(diagnostics []string, status int, secrets []string, extra ...string) *CheckResult {
	return accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, withStatus(diagnostics, status, extra...)), secrets...)
}

func forbiddenResult(diagnostics []string, status int, secrets []string, extra ...string) *CheckResult {
	return accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonProbeForbidden, core.AccessCheckItemStatusPassed, withStatus(diagnostics, status, extra...)), secrets...)
}

func (c *checker) classify(status int, body []byte, diagnostics []string, cfg *checkConfig) *CheckResult {
	secrets := cfg.redact()
	var payload struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(body, &payload)
	switch strings.ToLower(strings.TrimSpace(payload.Error)) {
	case "invalid_grant", "invalid_client", "unauthorized_client", "invalid_token":
		return authResult(diagnostics, status, secrets, "oauth "+payload.Error)
	}
	if gated := gateStatus(status, diagnostics, secrets, true); gated != nil {
		return gated
	}
	return remoteResult(diagnostics, status, secrets, "no credential verdict")
}

func (c *checker) services(ctx context.Context, client *http.Client, cfg *checkConfig, token string, diagnostics []string) (*CheckResult, error) {
	u, err := c.endpoint("dns-master", "services")
	if err != nil {
		return nil, fmt.Errorf("%w: probe url is invalid", core.ErrAccessCheckConfigInvalid)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: probe url is invalid", core.ErrAccessCheckConfigInvalid)
	}
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	secrets := append(cfg.redact(), token)
	if err != nil {
		return accesschecker.Finish(err, 0, "", diagnostics, secrets...), nil
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	return accesschecker.Finish(nil, resp.StatusCode, "", diagnostics, secrets...), nil
}
