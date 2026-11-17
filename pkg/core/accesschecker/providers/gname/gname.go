package gname

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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

const apiBaseURL = "https://api.gname.com"

type checker struct {
	baseURL string
	now func() time.Time
}

type checkConfig struct {
	AppId string
	AppKey string
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
	return string(domain.AccessProviderTypeGname)
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

	var wire domain.AccessConfigForGname
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		AppId: strings.TrimSpace(wire.AppId),
		AppKey: strings.TrimSpace(wire.AppKey),
	}
	if cfg.AppId == "" {
		return nil, fmt.Errorf("%w: appId is required", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.AppKey == "" {
		return nil, fmt.Errorf("%w: appKey is required", core.ErrAccessCheckConfigInvalid)
	}

	return cfg, nil
}

func (c *checkConfig) redact() []string {
	return []string{c.AppId, c.AppKey}
}

func (c *checker) newProbe(ctx context.Context, cfg *checkConfig) (*http.Request, error) {
u, err := c.endpoint("api", "domain", "list")
if err != nil {
	return nil, err
}
values := url.Values{}
values.Set("appid", cfg.AppId)
values.Set("gntime", strconv.FormatInt(c.clock().Unix(), 10))
values.Set("limit", "1")
values.Set("page", "1")
sum := md5.Sum([]byte(values.Encode() + cfg.AppKey))
values.Set("gntoken", strings.ToUpper(hex.EncodeToString(sum[:])))
req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), strings.NewReader(values.Encode()))
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
	if gated := gateStatus(status, diagnostics, secrets, true); gated != nil {
		return gated
	}
	var payload struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return remoteResult(diagnostics, status, secrets, "no credential verdict")
	}
	if payload.Code == 1 {
		return passResult(diagnostics, status, secrets)
	}
	if payload.Code == -1 && gnameAuth(payload.Msg) {
		return authResult(diagnostics, status, secrets, "code -1")
	}
	return remoteResult(diagnostics, status, secrets, "code "+strconv.Itoa(payload.Code))
}

func gnameAuth(msg string) bool {
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "sign") || strings.Contains(lower, "token") || strings.Contains(lower, "appid") || strings.Contains(lower, "appkey") || strings.Contains(lower, "secret") {
		return true
	}
	return strings.Contains(msg, "签名") || strings.Contains(msg, "密钥") || strings.Contains(msg, "认证")
}
