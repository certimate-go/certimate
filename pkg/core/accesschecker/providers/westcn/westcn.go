package westcn

import (
	"bytes"
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

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/accesschecker"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
)

const apiBaseURL = "https://api.west.cn/api/v2"

type checker struct {
	baseURL string
	now func() time.Time
}

type checkConfig struct {
	Username string
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
	return string(domain.AccessProviderTypeWestcn)
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

	var wire domain.AccessConfigForWestcn
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		Username: strings.TrimSpace(wire.Username),
		ApiPassword: strings.TrimSpace(wire.ApiPassword),
	}
	if cfg.Username == "" {
		return nil, fmt.Errorf("%w: username is required", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.ApiPassword == "" {
		return nil, fmt.Errorf("%w: apiPassword is required", core.ErrAccessCheckConfigInvalid)
	}

	return cfg, nil
}

func (c *checkConfig) redact() []string {
	return []string{c.Username, c.ApiPassword}
}

func (c *checker) newProbe(ctx context.Context, cfg *checkConfig) (*http.Request, error) {
u, err := c.endpoint("domain")
if err != nil {
	return nil, err
}
if !strings.HasSuffix(u.Path, "/") {
	u.Path += "/"
}
millis := strconv.FormatInt(c.clock().UnixMilli(), 10)
q := u.Query()
q.Set("act", "getdomains")
q.Set("limit", "1")
q.Set("page", "1")
q.Set("time", millis)
q.Set("token", westToken(cfg.Username, cfg.ApiPassword, millis))
q.Set("username", cfg.Username)
u.RawQuery = q.Encode()
req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
if err != nil {
	return nil, err
}
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
		Result  int `json:"result"`
		ErrCode int `json:"errcode"`
	}
	if err := decodeWest(body, &payload); err != nil {
		return remoteResult(diagnostics, status, secrets, "no credential verdict")
	}
	switch {
	case payload.Result == http.StatusOK:
		return passResult(diagnostics, status, secrets)
	case payload.ErrCode == 10000:
		return authResult(diagnostics, status, secrets, "errcode 10000")
	case payload.ErrCode == 10002:
		return forbiddenResult(diagnostics, status, secrets, "errcode 10002")
	case payload.ErrCode == 10001:
		return remoteResult(diagnostics, status, secrets, "errcode 10001")
	default:
		if payload.ErrCode != 0 {
			return remoteResult(diagnostics, status, secrets, "errcode "+strconv.Itoa(payload.ErrCode))
		}
		return remoteResult(diagnostics, status, secrets, "no credential verdict")
	}
}

func westToken(username, password, millis string) string {
	sum := md5.Sum([]byte(username + password + millis))
	return hex.EncodeToString(sum[:])
}

func decodeWest(body []byte, dest any) error {
	if err := json.Unmarshal(body, dest); err == nil {
		return nil
	}
	decoded, err := io.ReadAll(transform.NewReader(bytes.NewReader(body), simplifiedchinese.GBK.NewDecoder()))
	if err != nil {
		return err
	}
	return json.Unmarshal(decoded, dest)
}
