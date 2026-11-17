package acmeca

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

const (
	eabNotVerifiedNote   = "acme directory reachable; external account binding credentials are validated only when creating an ACME account, so they were not verified"
	invalidDirectoryNote = "response was not a valid ACME directory"
)

type checker struct{}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeACMECA)
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

	diagnostics := []string{accesschecker.HostOf(cfg.Endpoint)}
	return accesschecker.Redact(fetchDirectory(ctx, cfg.Endpoint, diagnostics), cfg.EabKid, cfg.EabHmacKey), nil
}

type checkConfig struct {
	EabKid     string
	EabHmacKey string
	Endpoint   string
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForACMECA
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		EabKid:     strings.TrimSpace(wire.EabKid),
		EabHmacKey: wire.EabHmacKey,
		Endpoint:   strings.TrimSpace(wire.Endpoint),
	}
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("%w: endpoint is required", core.ErrAccessCheckConfigInvalid)
	}
	if u, err := url.Parse(cfg.Endpoint); err != nil || !u.IsAbs() || u.Host == "" {
		return nil, fmt.Errorf("%w: endpoint must be an absolute URL", core.ErrAccessCheckConfigInvalid)
	}
	return cfg, nil
}

func fetchDirectory(ctx context.Context, endpoint string, diagnostics []string) *core.AccessCheckResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return accesschecker.Finish(err, 0, "", diagnostics)
	}

	resp, err := accesschecker.HTTPClient().Do(req)
	if err != nil {
		return accesschecker.Finish(err, 0, "", diagnostics)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return accesschecker.Finish(nil, resp.StatusCode, "", diagnostics)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return accesschecker.Finish(err, 0, "", diagnostics)
	}

	if isAcmeDirectory(body) {
		return accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, append(append([]string{}, diagnostics...), eabNotVerifiedNote))
	}
	return accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, append(append([]string{}, diagnostics...), invalidDirectoryNote))
}

func isAcmeDirectory(body []byte) bool {
	var directory map[string]any
	if err := json.Unmarshal(body, &directory); err != nil {
		return false
	}
	_, hasNewNonce := directory["newNonce"]
	_, hasNewAccount := directory["newAccount"]
	return hasNewNonce && hasNewAccount
}
