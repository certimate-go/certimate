package ovhcloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ovh/go-ovh/ovh"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/accesschecker"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
)

const (
	authMethodApplication = "application"
	authMethodOAuth2      = "oauth2"
)

type checker struct{}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeOVHcloud)
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

	client, err := buildClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", core.ErrAccessCheckConfigInvalid, err)
	}
	client.Timeout = accesschecker.BudgetLeft(ctx)

	diagnostics := []string{accesschecker.HostOf(cfg.Endpoint)}
	status, code, notes, callErr := currentCredential(ctx, client)
	diagnostics = append(diagnostics, notes...)
	var secrets []string
	secrets = append(secrets, cfg.ApplicationKey, cfg.ApplicationSecret, cfg.ConsumerKey, cfg.ClientId, cfg.ClientSecret)
	return accesschecker.Finish(callErr, status, code, diagnostics, secrets...), nil
}

type checkConfig struct {
	Endpoint          string
	AuthMethod        string
	ApplicationKey    string
	ApplicationSecret string
	ConsumerKey       string
	ClientId          string
	ClientSecret      string
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForOVHcloud
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		Endpoint:          strings.TrimSpace(wire.Endpoint),
		AuthMethod:        strings.TrimSpace(wire.AuthMethod),
		ApplicationKey:    strings.TrimSpace(wire.ApplicationKey),
		ApplicationSecret: wire.ApplicationSecret,
		ConsumerKey:       strings.TrimSpace(wire.ConsumerKey),
		ClientId:          strings.TrimSpace(wire.ClientId),
		ClientSecret:      wire.ClientSecret,
	}
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("%w: endpoint is required", core.ErrAccessCheckConfigInvalid)
	}

	switch cfg.AuthMethod {
	case authMethodApplication:
		if cfg.ApplicationKey == "" {
			return nil, fmt.Errorf("%w: applicationKey is required", core.ErrAccessCheckConfigInvalid)
		}
		if strings.TrimSpace(cfg.ApplicationSecret) == "" {
			return nil, fmt.Errorf("%w: applicationSecret is required", core.ErrAccessCheckConfigInvalid)
		}
		if cfg.ConsumerKey == "" {
			return nil, fmt.Errorf("%w: consumerKey is required", core.ErrAccessCheckConfigInvalid)
		}
	case authMethodOAuth2:
		if cfg.ClientId == "" {
			return nil, fmt.Errorf("%w: clientId is required", core.ErrAccessCheckConfigInvalid)
		}
		if strings.TrimSpace(cfg.ClientSecret) == "" {
			return nil, fmt.Errorf("%w: clientSecret is required", core.ErrAccessCheckConfigInvalid)
		}
	default:
		return nil, fmt.Errorf("%w: unsupported auth method %q", core.ErrAccessCheckConfigInvalid, cfg.AuthMethod)
	}
	return cfg, nil
}

func currentCredential(ctx context.Context, client *ovh.Client) (int, string, []string, error) {
	var credential map[string]any
	callErr := client.GetWithContext(ctx, "/auth/currentCredential", &credential)
	if callErr == nil {
		return http.StatusOK, "", nil, nil
	}

	var apiError *ovh.APIError
	if errors.As(callErr, &apiError) {
		return apiError.Code, apiError.Class, nil, nil
	}
	if credentialFault(callErr.Error()) {
		return http.StatusUnauthorized, "", nil, nil
	}
	return 0, "", nil, callErr
}

func buildClient(cfg *checkConfig) (*ovh.Client, error) {
	if cfg.AuthMethod == authMethodOAuth2 {
		return ovh.NewOAuth2Client(cfg.Endpoint, cfg.ClientId, cfg.ClientSecret)
	}
	return ovh.NewClient(cfg.Endpoint, cfg.ApplicationKey, cfg.ApplicationSecret, cfg.ConsumerKey)
}

func credentialFault(detail string) bool {
	detail = strings.ToLower(detail)
	for _, keyword := range []string{"invalid_client", "invalid_grant", "unauthorized_client"} {
		if strings.Contains(detail, keyword) {
			return true
		}
	}
	return false
}
