package oraclecloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	ocicommon "github.com/nrdcg/oci-go-sdk/common/v1065"
	ocicommonauth "github.com/nrdcg/oci-go-sdk/common/v1065/auth"
	ocidns "github.com/nrdcg/oci-go-sdk/dns/v1065"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/accesschecker"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
)

const (
	authMethodAPIKey            = "apikey"
	authMethodInstancePrincipal = "instanceprincipal"
	authMethodResourcePrincipal = "resourceprincipal"
	apiRegion                   = "us-ashburn-1"
)

type checker struct {
	endpoint string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeOracleCloud)
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
	if c.endpoint != "" {
		client.Host = c.endpoint
	}

	diagnostics := []string{accesschecker.HostOf(client.Host)}
	status, code, notes, callErr := listZones(ctx, client, cfg.TenancyOcid)
	diagnostics = append(diagnostics, notes...)
	return accesschecker.Finish(callErr, status, code, diagnostics, cfg.PrivateKey, cfg.PrivateKeyPassphrase), nil
}

type checkConfig struct {
	AuthMethod           string
	PrivateKey           string
	PrivateKeyPassphrase string
	PublicKeyFingerprint string
	TenancyOcid          string
	UserOcid             string
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForOracleCloud
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		AuthMethod:           strings.TrimSpace(wire.AuthMethod),
		PrivateKey:           wire.PrivateKey,
		PrivateKeyPassphrase: wire.PrivateKeyPassphrase,
		PublicKeyFingerprint: strings.TrimSpace(wire.PublicKeyFingerprint),
		TenancyOcid:          strings.TrimSpace(wire.TenancyOcid),
		UserOcid:             strings.TrimSpace(wire.UserOcid),
	}

	switch cfg.AuthMethod {
	case "", authMethodAPIKey:
		if strings.TrimSpace(cfg.PrivateKey) == "" {
			return nil, fmt.Errorf("%w: privateKey is required", core.ErrAccessCheckConfigInvalid)
		}
		if cfg.PublicKeyFingerprint == "" {
			return nil, fmt.Errorf("%w: publicKeyFingerprint is required", core.ErrAccessCheckConfigInvalid)
		}
		if cfg.TenancyOcid == "" {
			return nil, fmt.Errorf("%w: tenancyOcid is required", core.ErrAccessCheckConfigInvalid)
		}
		if cfg.UserOcid == "" {
			return nil, fmt.Errorf("%w: userOcid is required", core.ErrAccessCheckConfigInvalid)
		}
	case authMethodInstancePrincipal, authMethodResourcePrincipal:
		if cfg.TenancyOcid == "" {
			return nil, fmt.Errorf("%w: tenancyOcid is required", core.ErrAccessCheckConfigInvalid)
		}
	default:
		return nil, fmt.Errorf("%w: unsupported auth method %q", core.ErrAccessCheckConfigInvalid, cfg.AuthMethod)
	}
	return cfg, nil
}

func buildClient(cfg *checkConfig) (ocidns.DnsClient, error) {
	switch cfg.AuthMethod {
	case authMethodInstancePrincipal:
		provider, err := ocicommonauth.InstancePrincipalConfigurationProviderForRegion(ocicommon.Region(apiRegion))
		if err != nil {
			return ocidns.DnsClient{}, err
		}
		return ocidns.NewDnsClientWithConfigurationProvider(provider)
	case authMethodResourcePrincipal:
		provider, err := ocicommonauth.ResourcePrincipalConfigurationProviderForRegion(ocicommon.Region(apiRegion))
		if err != nil {
			return ocidns.DnsClient{}, err
		}
		return ocidns.NewDnsClientWithConfigurationProvider(provider)
	default:
		passphrase := (*string)(nil)
		if cfg.PrivateKeyPassphrase != "" {
			passphrase = &cfg.PrivateKeyPassphrase
		}
		provider := ocicommon.NewRawConfigurationProvider(
			cfg.TenancyOcid,
			cfg.UserOcid,
			apiRegion,
			cfg.PublicKeyFingerprint,
			cfg.PrivateKey,
			passphrase,
		)
		return ocidns.NewDnsClientWithConfigurationProvider(provider)
	}
}

func listZones(ctx context.Context, client ocidns.DnsClient, compartmentId string) (int, string, []string, error) {
	limit := int64(1)
	request := ocidns.ListZonesRequest{
		CompartmentId: &compartmentId,
		Limit:         &limit,
		RequestMetadata: ocicommon.RequestMetadata{
			RetryPolicy: func() *ocicommon.RetryPolicy { policy := ocicommon.NoRetryPolicy(); return &policy }(),
		},
	}

	_, callErr := client.ListZones(ctx, request)
	if callErr == nil {
		return http.StatusOK, "", nil, nil
	}

	if failure, ok := ocicommon.IsServiceError(callErr); ok {
		notes := []string{}
		if message := failure.GetMessage(); message != "" {
			notes = append(notes, message)
		}
		return failure.GetHTTPStatusCode(), failure.GetCode(), notes, nil
	}
	return 0, "", nil, callErr
}
