package namecheap

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net"
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
	defaultAPIURL = "https://api.namecheap.com/xml.response"
	defaultIPURL  = "https://dynamicdns.park-your-domain.com/getip"
)

const (
	errAPIKeyInvalid = "1011102"
	errAPIUser       = "1017105"
	errInvalidIP     = "1011150"
)

type checker struct {
	apiURL string
	ipURL  string
}

type checkConfig struct {
	Username string
	ApiKey   string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{apiURL: defaultAPIURL, ipURL: defaultIPURL}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeNamecheap)
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

	ipURL, err := parseProbeURL(c.ipURL, defaultIPURL)
	if err != nil {
		return nil, fmt.Errorf("%w: probe url is invalid", core.ErrAccessCheckConfigInvalid)
	}
	apiURL, err := parseProbeURL(c.apiURL, defaultAPIURL)
	if err != nil {
		return nil, fmt.Errorf("%w: probe url is invalid", core.ErrAccessCheckConfigInvalid)
	}

	client := accesschecker.HTTPClient()
	ip, result := c.lookupIP(ctx, client, ipURL, cfg)
	if result != nil {
		return result, nil
	}

	req, err := apiRequest(ctx, apiURL, cfg, ip)
	if err != nil {
		return nil, fmt.Errorf("%w: probe url is invalid", core.ErrAccessCheckConfigInvalid)
	}

	diagnostics := []string{apiURL.Host}
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

	return classifyNamecheap(resp.StatusCode, body, diagnostics, cfg), nil
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForNamecheap
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		Username: strings.TrimSpace(wire.Username),
		ApiKey:   strings.TrimSpace(wire.ApiKey),
	}
	if cfg.Username == "" {
		return nil, fmt.Errorf("%w: username is required", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.ApiKey == "" {
		return nil, fmt.Errorf("%w: apiKey is required", core.ErrAccessCheckConfigInvalid)
	}
	return cfg, nil
}

func (c *checkConfig) redact() []string {
	return []string{c.Username, c.ApiKey}
}

func parseProbeURL(raw, fallback string) (*url.URL, error) {
	if strings.TrimSpace(raw) == "" {
		raw = fallback
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, fmt.Errorf("probe url is invalid")
	}
	return u, nil
}

func (c *checker) lookupIP(ctx context.Context, client *http.Client, ipURL *url.URL, cfg *checkConfig) (string, *CheckResult) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ipURL.String(), nil)
	if err != nil {
		return "", accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusSkipped, []string{ipURL.Host, "client ip unreadable"}), cfg.redact()...)
	}

	diagnostics := []string{ipURL.Host}
	resp, err := client.Do(req)
	if err != nil {
		return "", accesschecker.Finish(err, 0, "", diagnostics, cfg.redact()...)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, append(diagnostics, fmt.Sprintf("HTTP %d", resp.StatusCode), "response body unreadable")), cfg.redact()...)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", accesschecker.Finish(nil, resp.StatusCode, "", diagnostics, cfg.redact()...)
	}

	ip := strings.TrimSpace(string(body))
	if net.ParseIP(ip) == nil {
		return "", accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, append(diagnostics, fmt.Sprintf("HTTP %d", resp.StatusCode), "client ip unreadable")), cfg.redact()...)
	}
	return ip, nil
}

func apiRequest(ctx context.Context, apiURL *url.URL, cfg *checkConfig, clientIP string) (*http.Request, error) {
	q := apiURL.Query()
	q.Set("ApiUser", cfg.Username)
	q.Set("ApiKey", cfg.ApiKey)
	q.Set("UserName", cfg.Username)
	q.Set("Command", "namecheap.domains.getList")
	q.Set("ClientIp", clientIP)
	q.Set("Page", "1")
	q.Set("PageSize", "10")
	endpoint := *apiURL
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/xml")
	return req, nil
}

type namecheapResponse struct {
	Status string `xml:"Status,attr"`
	Errors []struct {
		Number string `xml:"Number,attr"`
	} `xml:"Errors>Error"`
}

func classifyNamecheap(status int, body []byte, diagnostics []string, cfg *checkConfig) *CheckResult {
	secrets := cfg.redact()
	if status == http.StatusUnauthorized || status == http.StatusForbidden || status >= 500 || (status >= 300 && status < 400) {
		return accesschecker.Finish(nil, status, "", diagnostics, secrets...)
	}

	var payload namecheapResponse
	if err := xml.Unmarshal(body, &payload); err != nil {
		return remoteNamecheap(diagnostics, status, secrets, "no credential verdict")
	}

	switch strings.ToUpper(strings.TrimSpace(payload.Status)) {
	case "OK":
		diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", status))
		return accesschecker.Redact(accesschecker.Passed(diag), secrets...)
	case "ERROR":
		for _, item := range payload.Errors {
			switch strings.TrimSpace(item.Number) {
			case errAPIKeyInvalid, errAPIUser:
				return authNamecheap(diagnostics, status, secrets, "error "+item.Number)
			case errInvalidIP:
				return forbiddenNamecheap(diagnostics, status, secrets, "error "+item.Number)
			}
		}
		number := "empty"
		if len(payload.Errors) > 0 && strings.TrimSpace(payload.Errors[0].Number) != "" {
			number = strings.TrimSpace(payload.Errors[0].Number)
		}
		return remoteNamecheap(diagnostics, status, secrets, "error "+number)
	default:
		return remoteNamecheap(diagnostics, status, secrets, "no credential verdict")
	}
}

func authNamecheap(diagnostics []string, status int, secrets []string, extra string) *CheckResult {
	diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", status), extra)
	return accesschecker.Redact(accesschecker.AuthRejected(diag), secrets...)
}

func forbiddenNamecheap(diagnostics []string, status int, secrets []string, extra string) *CheckResult {
	diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", status), extra)
	return accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonProbeForbidden, core.AccessCheckItemStatusPassed, diag), secrets...)
}

func remoteNamecheap(diagnostics []string, status int, secrets []string, extra string) *CheckResult {
	diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", status), extra)
	return accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, diag), secrets...)
}
