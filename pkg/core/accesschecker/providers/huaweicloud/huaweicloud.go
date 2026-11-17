package huaweicloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	hwauth "github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	hwconfig "github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdkerr"
	hwdns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"
	hwmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
	hwregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/region"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/accesschecker"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
)

type checker struct {
	endpoint string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeHuaweiCloud)
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

	host := "dns.cn-north-4.myhuaweicloud.com"
	if c.endpoint != "" {
		host = accesschecker.HostOf(c.endpoint)
	}
	diagnostics := []string{host}
	status, code, callErr := listPublicZones(ctx, c.endpoint, cfg.AccessKeyId, cfg.SecretAccessKey)
	return accesschecker.Finish(callErr, status, code, diagnostics, cfg.AccessKeyId, cfg.SecretAccessKey), nil
}

type checkConfig struct {
	AccessKeyId     string
	SecretAccessKey string
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForHuaweiCloud
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		AccessKeyId:     strings.TrimSpace(wire.AccessKeyId),
		SecretAccessKey: wire.SecretAccessKey,
	}
	if cfg.AccessKeyId == "" {
		return nil, fmt.Errorf("%w: accessKeyId is required", core.ErrAccessCheckConfigInvalid)
	}
	if strings.TrimSpace(cfg.SecretAccessKey) == "" {
		return nil, fmt.Errorf("%w: secretAccessKey is required", core.ErrAccessCheckConfigInvalid)
	}
	return cfg, nil
}

func listPublicZones(ctx context.Context, endpoint, accessKeyId, secretAccessKey string) (int, string, error) {
	auth, err := hwauth.NewCredentialsBuilder().
		WithAk(accessKeyId).
		WithSk(secretAccessKey).
		SafeBuild()
	if err != nil {
		return 0, "", err
	}

	httpCfg := hwconfig.DefaultHttpConfig().
		WithTimeout(accesschecker.BudgetLeft(ctx)).
		WithRetries(0)
	builder := hwdns.DnsClientBuilder().
		WithHttpConfig(httpCfg).
		WithCredential(auth)
	if endpoint != "" {
		builder = builder.WithEndpoint(endpoint)
	} else {
		builder = builder.WithRegion(hwregion.CN_NORTH_4)
	}
	hc, err := builder.SafeBuild()
	if err != nil {
		return 0, "", err
	}

	limit := int32(1)
	var callErr error
	err = accesschecker.Call(ctx, func() error {
		_, callErr = hwdns.NewDnsClient(hc).ListPublicZones(&hwmodel.ListPublicZonesRequest{Limit: &limit})
		return callErr
	})
	if err != nil {
		return sdkFault(err)
	}
	return 200, "", nil
}

func sdkFault(err error) (int, string, error) {
	var svcErr *sdkerr.ServiceResponseError
	if errors.As(err, &svcErr) && svcErr != nil {
		return svcErr.StatusCode, svcErr.ErrorCode, err
	}
	return 0, "", err
}
