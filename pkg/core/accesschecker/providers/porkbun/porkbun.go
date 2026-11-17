package porkbun

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

const apiBaseURL = "https://api.porkbun.com/api/json/v3"

type checker struct {
	baseURL string
}

type checkConfig struct {
	ApiKey string
	SecretApiKey string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{baseURL: apiBaseURL}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypePorkbun)
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
	resp, err := c.httpClient(cfg).Do(req)
	if err != nil {
		return accesschecker.Finish(err, 0, "", diagnostics, cfg.redact()...), nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", resp.StatusCode), "response body unreadable")
		return accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, diag), cfg.redact()...), nil
	}

	return c.classify(resp.StatusCode, body, diagnostics, cfg), nil
}

func (c *checker) httpClient(cfg *checkConfig) *http.Client {
	client := accesschecker.HTTPClient()
	_ = cfg
	return client
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForPorkbun
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		ApiKey: strings.TrimSpace(wire.ApiKey),
		SecretApiKey: strings.TrimSpace(wire.SecretApiKey),
	}
	if cfg.ApiKey == "" {
		return nil, fmt.Errorf("%w: apiKey is required", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.SecretApiKey == "" {
		return nil, fmt.Errorf("%w: secretApiKey is required", core.ErrAccessCheckConfigInvalid)
	}

	return cfg, nil
}

func (c *checkConfig) redact() []string {
	return []string{c.ApiKey, c.SecretApiKey}
}

func (c *checker) newProbe(ctx context.Context, cfg *checkConfig) (*http.Request, error) {
u, err := c.endpoint("ping")
if err != nil {
	return nil, err
}
payload, err := json.Marshal(map[string]string{
	"apikey":       cfg.ApiKey,
	"secretapikey": cfg.SecretApiKey,
})
if err != nil {
	return nil, err
}
req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), strings.NewReader(string(payload)))
if err != nil {
	return nil, err
}
req.Header.Set("Accept", "application/json")
req.Header.Set("Content-Type", "application/json")
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
	if gated := gateStatus(status, diagnostics, secrets, true); gated != nil {
		return gated
	}
	var payload struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return remoteResult(diagnostics, status, secrets, "no credential verdict")
	}
	switch strings.ToUpper(strings.TrimSpace(payload.Status)) {
	case "SUCCESS":
		return passResult(diagnostics, status, secrets)
	case "ERROR":
		text := strings.ToLower(payload.Message)
		if strings.Contains(text, "limit") || strings.Contains(text, "temporar") || strings.Contains(text, "unavailable") {
			return remoteResult(diagnostics, status, secrets, "status error")
		}
		return authResult(diagnostics, status, secrets, "status error")
	default:
		if status >= 200 && status < 300 {
			return remoteResult(diagnostics, status, secrets, "no credential verdict")
		}
		return accesschecker.Finish(nil, status, "", diagnostics, secrets...)
	}
}
