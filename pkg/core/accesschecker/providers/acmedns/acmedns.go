package acmedns

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

const notVerifiedNote = "server reachable; acme-dns credentials are validated only when updating records, so they were not verified"

type checker struct{}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeACMEDNS)
}

func (c *checker) Scope() []string {
	return []string{core.AccessCheckItemConnect, core.AccessCheckItemAuth}
}

type checkConfig struct {
	ServerUrl  *url.URL
	Credential []string
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForACMEDNS
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{}
	if strings.TrimSpace(wire.ServerUrl) == "" {
		return nil, fmt.Errorf("%w: serverUrl is required", core.ErrAccessCheckConfigInvalid)
	}
	cfg.ServerUrl, err = url.Parse(strings.TrimSpace(wire.ServerUrl))
	if err != nil || cfg.ServerUrl.Host == "" || (cfg.ServerUrl.Scheme != "http" && cfg.ServerUrl.Scheme != "https") {
		return nil, fmt.Errorf("%w: serverUrl must be a valid http or https URL", core.ErrAccessCheckConfigInvalid)
	}
	if strings.TrimSpace(wire.Credentials) == "" {
		return nil, fmt.Errorf("%w: credentials are required", core.ErrAccessCheckConfigInvalid)
	}
	cfg.Credential, err = collectSecrets(strings.TrimSpace(wire.Credentials))
	if err != nil {
		return nil, fmt.Errorf("%w: credentials must be a JSON object of acme-dns accounts", core.ErrAccessCheckConfigInvalid)
	}
	return cfg, nil
}

type acmednsAccount struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func collectSecrets(raw string) ([]string, error) {
	var accounts map[string]acmednsAccount
	if err := json.Unmarshal([]byte(raw), &accounts); err == nil {
		secrets := make([]string, 0, len(accounts)*2)
		for _, account := range accounts {
			secrets = append(secrets, account.Username, account.Password)
		}
		return secrets, nil
	}
	var account acmednsAccount
	if err := json.Unmarshal([]byte(raw), &account); err != nil {
		return nil, err
	}
	return []string{account.Username, account.Password}, nil
}

func (c *checker) Check(ctx context.Context, config map[string]any) (*CheckResult, error) {
	ctx, cancel := accesschecker.WithBudget(ctx)
	defer cancel()

	cfg, err := parseConfig(config)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.ServerUrl.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: serverUrl is invalid", core.ErrAccessCheckConfigInvalid)
	}
	req.Header.Set("Accept", "application/json")

	diagnostics := []string{cfg.ServerUrl.Host}
	resp, err := accesschecker.HTTPClient().Do(req)
	if err != nil {
		if result, ok := accesschecker.FromTransport(err, diagnostics); ok {
			return accesschecker.Redact(result, cfg.Credential...), nil
		}
		return accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, append(append([]string{}, diagnostics...), err.Error())), cfg.Credential...), nil
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))

	diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", resp.StatusCode), notVerifiedNote)
	return accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, diag), cfg.Credential...), nil
}
