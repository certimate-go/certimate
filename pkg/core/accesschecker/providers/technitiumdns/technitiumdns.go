package technitiumdns

import (
	"context"
	"crypto/tls"
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

type checker struct{}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeTechnitiumDNS)
}

func (c *checker) Scope() []string {
	return []string{core.AccessCheckItemConnect, core.AccessCheckItemAuth}
}

type checkConfig struct {
	Endpoint                 *url.URL
	ApiToken                 string
	AllowInsecureConnections bool
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForTechnitiumDNS
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		ApiToken:                 strings.TrimSpace(wire.ApiToken),
		AllowInsecureConnections: wire.AllowInsecureConnections,
	}
	if strings.TrimSpace(wire.ServerUrl) == "" {
		return nil, fmt.Errorf("%w: serverUrl is required", core.ErrAccessCheckConfigInvalid)
	}
	cfg.Endpoint, err = url.Parse(strings.TrimSpace(wire.ServerUrl))
	if err != nil || cfg.Endpoint.Host == "" || (cfg.Endpoint.Scheme != "http" && cfg.Endpoint.Scheme != "https") {
		return nil, fmt.Errorf("%w: serverUrl must be a valid http or https URL", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.ApiToken == "" {
		return nil, fmt.Errorf("%w: apiToken is required", core.ErrAccessCheckConfigInvalid)
	}
	return cfg, nil
}

func (c *checker) Check(ctx context.Context, config map[string]any) (*CheckResult, error) {
	ctx, cancel := accesschecker.WithBudget(ctx)
	defer cancel()

	cfg, err := parseConfig(config)
	if err != nil {
		return nil, err
	}

	endpoint := cfg.Endpoint.JoinPath("api", "zones", "list")

	form := url.Values{}
	form.Set("token", cfg.ApiToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: probe url is invalid", core.ErrAccessCheckConfigInvalid)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	diagnostics := []string{endpoint.Host}
	resp, err := httpClient(cfg).Do(req)
	if err != nil {
		if result, ok := accesschecker.FromTransport(err, diagnostics); ok {
			return accesschecker.Redact(result, cfg.ApiToken), nil
		}
		return accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, append(append([]string{}, diagnostics...), err.Error())), cfg.ApiToken), nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", resp.StatusCode), "response body unreadable")
		return accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, diag), cfg.ApiToken), nil
	}

	return classify(resp.StatusCode, body, diagnostics, cfg.ApiToken), nil
}

func httpClient(cfg *checkConfig) *http.Client {
	client := accesschecker.HTTPClient()
	if cfg.AllowInsecureConnections {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		client.Transport = transport
	}
	return client
}

func classify(status int, body []byte, diagnostics []string, apiToken string) *CheckResult {
	var payload struct {
		Status       string `json:"status"`
		ErrorMessage string `json:"errorMessage"`
	}
	_ = json.Unmarshal(body, &payload)

	diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", status))
	switch {
	case status == http.StatusUnauthorized:
		return accesschecker.Redact(accesschecker.AuthRejected(diag), apiToken)
	case status == http.StatusForbidden:
		return accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonProbeForbidden, core.AccessCheckItemStatusPassed, diag), apiToken)
	case status < http.StatusBadRequest && payload.Status == "ok":
		return accesschecker.Redact(accesschecker.Passed(diag), apiToken)
	case status < http.StatusBadRequest && payload.Status == "invalid-token":
		return accesschecker.Redact(accesschecker.AuthRejected(append(diag, "status invalid-token")), apiToken)
	}

	diag = append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", status))
	if payload.Status != "" {
		diag = append(diag, "status "+payload.Status)
	}
	if payload.ErrorMessage != "" {
		diag = append(diag, payload.ErrorMessage)
	}
	if payload.Status == "" && payload.ErrorMessage == "" {
		diag = append(diag, "no credential verdict")
	}
	return accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, diag), apiToken)
}
