package xinnet

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/accesschecker"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
)

const apiBaseURL = "https://apiv2.xinnet.com"

type checker struct {
	baseURL string
	now     func() time.Time
}

type checkConfig struct {
	AgentId     string
	ApiPassword string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{baseURL: apiBaseURL, now: time.Now}
}

func (c *checker) clock() time.Time {
	if c.now != nil {
		return c.now()
	}
	return time.Now()
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeXinnet)
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

	var wire domain.AccessConfigForXinnet
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		AgentId:     strings.TrimSpace(wire.AgentId),
		ApiPassword: strings.TrimSpace(wire.ApiPassword),
	}
	if cfg.AgentId == "" {
		return nil, fmt.Errorf("%w: agentId is required", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.ApiPassword == "" {
		return nil, fmt.Errorf("%w: apiPassword is required", core.ErrAccessCheckConfigInvalid)
	}

	return cfg, nil
}

func (c *checkConfig) redact() []string {
	return []string{c.AgentId, c.ApiPassword}
}

func (c *checker) newProbe(ctx context.Context, cfg *checkConfig) (*http.Request, error) {
	u, err := c.endpoint("api", "domain", "list")
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(struct {
		PageNo   int `json:"pageNo"`
		PageSize int `json:"pageSize"`
	}{PageNo: 1, PageSize: 1})
	if err != nil {
		return nil, err
	}
	timestamp := c.clock().UTC().Format("20060102T150405Z")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), strings.NewReader(string(raw)))
	if err != nil {
		return nil, err
	}
	signedPath := req.URL.Path
	if !strings.HasSuffix(signedPath, "/") {
		signedPath += "/"
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Timestamp", timestamp)
	req.Header.Set("Authorization", "HMAC-SHA256 Access="+cfg.AgentId+", Signature="+signXinnet(cfg.ApiPassword, timestamp, http.MethodPost, signedPath, string(raw)))
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
		Code json.RawMessage `json:"code"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return remoteResult(diagnostics, status, secrets, "no credential verdict")
	}
	code := strings.Trim(strings.TrimSpace(string(payload.Code)), `"`)
	switch code {
	case "0":
		return passResult(diagnostics, status, secrets)
	case "350002", "350003":
		return authResult(diagnostics, status, secrets, "code "+code)
	case "355020":
		return forbiddenResult(diagnostics, status, secrets, "code "+code)
	default:
		if code == "" || code == "null" {
			return remoteResult(diagnostics, status, secrets, "no credential verdict")
		}
		return remoteResult(diagnostics, status, secrets, "code "+code)
	}
}

func signXinnet(secret, timestamp, method, urlPath, body string) string {
	stringToSign := "HMAC-SHA256\n" + timestamp + "\n" + method + "\n" + urlPath + "\n" + body
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(stringToSign))
	return hex.EncodeToString(mac.Sum(nil))
}
