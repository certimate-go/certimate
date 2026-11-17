package ionos

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

const apiBaseURL = "https://api.hosting.ionos.com/dns"

type checker struct {
	baseURL string
}

type checkConfig struct {
	ApiKeyPublicPrefix string
	ApiKeySecret string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{baseURL: apiBaseURL}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeIONOS)
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

	var wire domain.AccessConfigForIONOS
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		ApiKeyPublicPrefix: strings.TrimSpace(wire.ApiKeyPublicPrefix),
		ApiKeySecret: strings.TrimSpace(wire.ApiKeySecret),
	}
	if cfg.ApiKeyPublicPrefix == "" {
		return nil, fmt.Errorf("%w: apiKeyPublicPrefix is required", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.ApiKeySecret == "" {
		return nil, fmt.Errorf("%w: apiKeySecret is required", core.ErrAccessCheckConfigInvalid)
	}

	return cfg, nil
}

func (c *checkConfig) redact() []string {
	return []string{c.ApiKeyPublicPrefix, c.ApiKeySecret}
}

func (c *checker) newProbe(ctx context.Context, cfg *checkConfig) (*http.Request, error) {
u, err := c.endpoint("v1", "zones")
if err != nil {
	return nil, err
}
req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
if err != nil {
	return nil, err
}
req.Header.Set("Accept", "application/json")
req.Header.Set("X-API-Key", cfg.ApiKeyPublicPrefix+"."+cfg.ApiKeySecret)
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
