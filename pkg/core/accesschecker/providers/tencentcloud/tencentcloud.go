package tencentcloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	tccommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/accesschecker"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
)

const (
	apiHost   = "dnspod.tencentcloudapi.com"
	apiScheme = "HTTPS"
	apiRegion = "ap-guangzhou"
)

type checker struct {
	endpoint string
	scheme   string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{endpoint: apiHost, scheme: apiScheme}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeTencentCloud)
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

	endpoint := c.endpoint
	scheme := c.scheme
	if endpoint == "" {
		endpoint = apiHost
	}
	if scheme == "" {
		scheme = apiScheme
	}

	diagnostics := []string{endpoint}
	status, code, callErr := describeDomainList(ctx, endpoint, scheme, cfg.SecretId, cfg.SecretKey)
	return accesschecker.Finish(callErr, status, code, diagnostics, cfg.SecretId, cfg.SecretKey), nil
}

type checkConfig struct {
	SecretId  string
	SecretKey string
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForTencentCloud
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		SecretId:  strings.TrimSpace(wire.SecretId),
		SecretKey: wire.SecretKey,
	}
	if cfg.SecretId == "" {
		return nil, fmt.Errorf("%w: secretId is required", core.ErrAccessCheckConfigInvalid)
	}
	if strings.TrimSpace(cfg.SecretKey) == "" {
		return nil, fmt.Errorf("%w: secretKey is required", core.ErrAccessCheckConfigInvalid)
	}
	return cfg, nil
}

func describeDomainList(ctx context.Context, endpoint, scheme, secretId, secretKey string) (int, string, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = endpoint
	cpf.HttpProfile.Scheme = scheme
	cpf.DisableRegionBreaker = true
	seconds := int(accesschecker.BudgetLeft(ctx) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	cpf.HttpProfile.ReqTimeout = seconds

	client := tccommon.NewCommonClient(tccommon.NewCredential(secretId, secretKey), apiRegion, cpf)
	request := tchttp.NewCommonRequest("dnspod", "2021-03-23", "DescribeDomainList")
	if err := request.SetActionParameters(map[string]any{"Offset": 0, "Limit": 1}); err != nil {
		return 0, "", err
	}
	response := tchttp.NewCommonResponse()
	err := accesschecker.Call(ctx, func() error {
		return client.Send(request, response)
	})
	if err != nil {
		return sdkFault(err)
	}
	return 200, "", nil
}

func sdkFault(err error) (int, string, error) {
	var sdkErr *tcerr.TencentCloudSDKError
	if errors.As(err, &sdkErr) && sdkErr != nil {
		return 0, sdkErr.Code, err
	}
	return 0, "", err
}
