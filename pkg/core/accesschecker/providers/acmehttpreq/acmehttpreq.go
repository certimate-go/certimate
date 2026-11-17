package acmehttpreq

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

const noCredentialsNote = "no credentials configured; connectivity only"

type checker struct{}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeACMEHttpReq)
}

func (c *checker) Scope() []string {
	return []string{core.AccessCheckItemConnect, core.AccessCheckItemAuth}
}

type checkConfig struct {
	Endpoint     *url.URL
	Username     string
	Password     string
	HasPassword  bool
	PartialCreds bool
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForACMEHttpReq
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		Username: strings.TrimSpace(wire.Username),
		Password: wire.Password,
	}
	if strings.TrimSpace(wire.Endpoint) == "" {
		return nil, fmt.Errorf("%w: endpoint is required", core.ErrAccessCheckConfigInvalid)
	}
	cfg.Endpoint, err = url.Parse(strings.TrimSpace(wire.Endpoint))
	if err != nil || cfg.Endpoint.Host == "" || (cfg.Endpoint.Scheme != "http" && cfg.Endpoint.Scheme != "https") {
		return nil, fmt.Errorf("%w: endpoint must be a valid http or https URL", core.ErrAccessCheckConfigInvalid)
	}
	cfg.HasPassword = cfg.Username != "" && strings.TrimSpace(cfg.Password) != ""
	cfg.PartialCreds = !cfg.HasPassword && (cfg.Username != "" || strings.TrimSpace(cfg.Password) != "")
	return cfg, nil
}

func (c *checker) Check(ctx context.Context, config map[string]any) (*CheckResult, error) {
	ctx, cancel := accesschecker.WithBudget(ctx)
	defer cancel()

	cfg, err := parseConfig(config)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.Endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: endpoint is invalid", core.ErrAccessCheckConfigInvalid)
	}
	req.Header.Set("Accept", "application/json")
	if cfg.HasPassword {
		req.SetBasicAuth(cfg.Username, cfg.Password)
	}

	diagnostics := []string{cfg.Endpoint.Host}
	resp, err := accesschecker.HTTPClient().Do(req)
	if err != nil {
		if result, ok := accesschecker.FromTransport(err, diagnostics); ok {
			return accesschecker.Redact(result, cfg.Username, cfg.Password), nil
		}
		return accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, append(append([]string{}, diagnostics...), err.Error())), cfg.Username, cfg.Password), nil
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))

	diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", resp.StatusCode))
	if cfg.HasPassword {
		return accesschecker.Finish(nil, resp.StatusCode, "", diag, cfg.Username, cfg.Password), nil
	}
	if cfg.PartialCreds {
		diag = append(diag, "credentials are incomplete and were not sent")
	}
	diag = append(diag, noCredentialsNote)
	return accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, diag), cfg.Username, cfg.Password), nil
}
