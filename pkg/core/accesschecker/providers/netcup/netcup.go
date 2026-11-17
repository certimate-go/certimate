package netcup

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

const apiBaseURL = "https://ccp.netcup.net/run/webservice/servers/endpoint.php?JSON"

type checker struct {
	baseURL string
}

type checkConfig struct {
	CustomerNumber string
	ApiKey string
	ApiPassword string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{baseURL: apiBaseURL}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeNetcup)
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
	resp, err := accesschecker.HTTPClient().Do(req)
	if err != nil {
		return accesschecker.Finish(err, 0, "", diagnostics, cfg.redact()...), nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", resp.StatusCode), "response body unreadable")
		return accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, diag), cfg.redact()...), nil
	}

	result := c.classify(resp.StatusCode, body, diagnostics, cfg)
	if result.Status == core.AccessCheckStatusPassed {
		c.logout(ctx, cfg, sessionID(body))
	}
	return result, nil
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForNetcup
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		CustomerNumber: strings.TrimSpace(wire.CustomerNumber),
		ApiKey: strings.TrimSpace(wire.ApiKey),
		ApiPassword: strings.TrimSpace(wire.ApiPassword),
	}
	if cfg.CustomerNumber == "" {
		return nil, fmt.Errorf("%w: customerNumber is required", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.ApiKey == "" {
		return nil, fmt.Errorf("%w: apiKey is required", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.ApiPassword == "" {
		return nil, fmt.Errorf("%w: apiPassword is required", core.ErrAccessCheckConfigInvalid)
	}

	return cfg, nil
}

func (c *checkConfig) redact() []string {
	return []string{c.CustomerNumber, c.ApiKey, c.ApiPassword}
}

func (c *checker) newProbe(ctx context.Context, cfg *checkConfig) (*http.Request, error) {
return c.post(ctx, cfg, "login", loginParam{CustomerNumber: cfg.CustomerNumber, APIKey: cfg.ApiKey, APIPassword: cfg.ApiPassword})
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
		Status       string `json:"status"`
		ShortMessage string `json:"shortmessage"`
		LongMessage  string `json:"longmessage"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return remoteResult(diagnostics, status, secrets, "no credential verdict")
	}
	if strings.EqualFold(strings.TrimSpace(payload.Status), "success") {
		return passResult(diagnostics, status, secrets)
	}
	text := strings.ToLower(payload.ShortMessage + " " + payload.LongMessage)
	if strings.Contains(text, "limit") || strings.Contains(text, "temporar") || strings.Contains(text, "maintenance") {
		return remoteResult(diagnostics, status, secrets, "status "+payload.Status)
	}
	if strings.EqualFold(strings.TrimSpace(payload.Status), "error") {
		return authResult(diagnostics, status, secrets, "status error")
	}
	return remoteResult(diagnostics, status, secrets, "no credential verdict")
}

type loginParam struct {
	CustomerNumber string `json:"customernumber"`
	APIKey         string `json:"apikey"`
	APIPassword    string `json:"apipassword"`
}

type logoutParam struct {
	CustomerNumber string `json:"customernumber"`
	APIKey         string `json:"apikey"`
	APISessionID   string `json:"apisessionid"`
}

func (c *checker) post(ctx context.Context, cfg *checkConfig, action string, param any) (*http.Request, error) {
	u, err := c.endpoint()
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(struct {
		Action string `json:"action"`
		Param  any    `json:"param"`
	}{Action: action, Param: param})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), strings.NewReader(string(raw)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func sessionID(body []byte) string {
	var payload struct {
		ResponseData struct {
			APISessionID string `json:"apisessionid"`
		} `json:"responsedata"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	return payload.ResponseData.APISessionID
}

func (c *checker) logout(ctx context.Context, cfg *checkConfig, session string) {
	if session == "" {
		return
	}
	req, err := c.post(ctx, cfg, "logout", logoutParam{CustomerNumber: cfg.CustomerNumber, APIKey: cfg.ApiKey, APISessionID: session})
	if err != nil {
		return
	}
	resp, err := accesschecker.HTTPClient().Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
}
