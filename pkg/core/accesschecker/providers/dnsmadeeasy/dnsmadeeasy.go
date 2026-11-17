package dnsmadeeasy

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
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

const apiBaseURL = "https://api.dnsmadeeasy.com/V2.0"

type checker struct {
	baseURL string
	now func() time.Time
}

type checkConfig struct {
	ApiKey string
	ApiSecret string
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
	return string(domain.AccessProviderTypeDNSMadeEasy)
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

	var wire domain.AccessConfigForDNSMadeEasy
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		ApiKey: strings.TrimSpace(wire.ApiKey),
		ApiSecret: strings.TrimSpace(wire.ApiSecret),
	}
	if cfg.ApiKey == "" {
		return nil, fmt.Errorf("%w: apiKey is required", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.ApiSecret == "" {
		return nil, fmt.Errorf("%w: apiSecret is required", core.ErrAccessCheckConfigInvalid)
	}

	return cfg, nil
}

func (c *checkConfig) redact() []string {
	return []string{c.ApiKey, c.ApiSecret}
}

func (c *checker) newProbe(ctx context.Context, cfg *checkConfig) (*http.Request, error) {
u, err := c.endpoint("dns", "managed")
if err != nil {
	return nil, err
}
if !strings.HasSuffix(u.Path, "/") {
	u.Path += "/"
}
date := c.clock().UTC().Format(time.RFC1123)
req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
if err != nil {
	return nil, err
}
req.Header.Set("Accept", "application/json")
req.Header.Set("x-dnsme-apiKey", cfg.ApiKey)
req.Header.Set("x-dnsme-requestDate", date)
req.Header.Set("x-dnsme-hmac", dnsmeHMAC(cfg.ApiSecret, date))
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

func (c *checker) classify(status int, _ []byte, diagnostics []string, cfg *checkConfig) *CheckResult {
	return accesschecker.Finish(nil, status, "", diagnostics, cfg.redact()...)
}

func dnsmeHMAC(secret, date string) string {
	mac := hmac.New(sha1.New, []byte(secret))
	_, _ = mac.Write([]byte(date))
	return hex.EncodeToString(mac.Sum(nil))
}
