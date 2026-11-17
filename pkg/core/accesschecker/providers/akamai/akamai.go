package akamai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v14/pkg/dns"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v14/pkg/edgegrid"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v14/pkg/session"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/accesschecker"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
)

const edgeDNSMaxBody = 131072

type checker struct {
	client *http.Client
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeAkamai)
}

func (c *checker) Scope() []string {
	return []string{core.AccessCheckItemConnect, core.AccessCheckItemAuth}
}

type checkConfig struct {
	Host         string
	ClientToken  string
	ClientSecret string
	AccessToken  string
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForAkamai
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		Host:         strings.TrimSpace(wire.Host),
		ClientToken:  strings.TrimSpace(wire.ClientToken),
		ClientSecret: strings.TrimSpace(wire.ClientSecret),
		AccessToken:  strings.TrimSpace(wire.AccessToken),
	}
	if cfg.Host == "" {
		return nil, fmt.Errorf("%w: host is required", core.ErrAccessCheckConfigInvalid)
	}
	if strings.ContainsAny(cfg.Host, "/ ") {
		return nil, fmt.Errorf("%w: host must be a hostname without scheme or path", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.ClientToken == "" {
		return nil, fmt.Errorf("%w: clientToken is required", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.ClientSecret == "" {
		return nil, fmt.Errorf("%w: clientSecret is required", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.AccessToken == "" {
		return nil, fmt.Errorf("%w: accessToken is required", core.ErrAccessCheckConfigInvalid)
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

	client := c.client
	if client == nil {
		base := accesschecker.HTTPClient()
		client = &(*base)
	}
	client.Timeout = accesschecker.BudgetLeft(ctx)

	sess, err := session.New(
		session.WithSigner(&edgegrid.Config{
			Host:         cfg.Host,
			ClientToken:  cfg.ClientToken,
			ClientSecret: cfg.ClientSecret,
			AccessToken:  cfg.AccessToken,
			MaxBody:      edgeDNSMaxBody,
		}),
		session.WithClient(client),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: edgegrid config is invalid", core.ErrAccessCheckConfigInvalid)
	}

	diagnostics := []string{cfg.Host}
	secrets := []string{cfg.ClientToken, cfg.ClientSecret, cfg.AccessToken}
	_, err = dns.Client(sess).ListZones(ctx, dns.ListZonesRequest{Page: 1, PageSize: 1})
	if err != nil {
		if result, ok := accesschecker.FromTransport(err, diagnostics); ok {
			return accesschecker.Redact(result, secrets...), nil
		}
		var dnsErr *dns.Error
		if errors.As(err, &dnsErr) {
			return accesschecker.Finish(nil, dnsErr.StatusCode, dnsErr.Title, diagnostics, secrets...), nil
		}
		diag := append(append([]string{}, diagnostics...), "no credential verdict", err.Error())
		return accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, diag), secrets...), nil
	}
	return accesschecker.Finish(nil, http.StatusOK, "", diagnostics, secrets...), nil
}
