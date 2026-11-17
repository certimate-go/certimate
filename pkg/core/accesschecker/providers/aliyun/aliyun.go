package aliyun

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	openapiutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/accesschecker"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
)

const (
	apiEndpoint = "alidns.aliyuncs.com"
	apiProtocol = "HTTPS"
	apiRegion   = "cn-hangzhou"
)

type checker struct {
	endpoint string
	protocol string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{endpoint: apiEndpoint, protocol: apiProtocol}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeAliyun)
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
	protocol := c.protocol
	if endpoint == "" {
		endpoint = apiEndpoint
	}
	if protocol == "" {
		protocol = apiProtocol
	}

	diagnostics := []string{endpoint}
	status, code, callErr := describeDomains(ctx, endpoint, protocol, cfg.AccessKeyId, cfg.AccessKeySecret)
	return accesschecker.Finish(callErr, status, code, diagnostics, cfg.AccessKeyId, cfg.AccessKeySecret), nil
}

type checkConfig struct {
	AccessKeyId     string
	AccessKeySecret string
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForAliyun
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		AccessKeyId:     strings.TrimSpace(wire.AccessKeyId),
		AccessKeySecret: wire.AccessKeySecret,
	}
	if cfg.AccessKeyId == "" {
		return nil, fmt.Errorf("%w: accessKeyId is required", core.ErrAccessCheckConfigInvalid)
	}
	if strings.TrimSpace(cfg.AccessKeySecret) == "" {
		return nil, fmt.Errorf("%w: accessKeySecret is required", core.ErrAccessCheckConfigInvalid)
	}
	return cfg, nil
}

func describeDomains(ctx context.Context, endpoint, protocol, accessKeyId, accessKeySecret string) (int, string, error) {
	client, err := openapi.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(endpoint),
		Protocol:        tea.String(protocol),
		RegionId:        tea.String(apiRegion),
	})
	if err != nil {
		return 0, "", err
	}

	timeoutMs := int(accesschecker.BudgetLeft(ctx) / time.Millisecond)
	if timeoutMs < 1 {
		timeoutMs = 1
	}
	_, err = client.CallApiWithCtx(ctx, &openapiutil.Params{
		Action:      dara.String("DescribeDomains"),
		Version:     dara.String("2015-01-09"),
		Protocol:    dara.String(protocol),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}, &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(map[string]any{
			"PageNumber": 1,
			"PageSize":   1,
		}),
	}, &dara.RuntimeOptions{
		Autoretry:      tea.Bool(false),
		MaxAttempts:    tea.Int(1),
		ReadTimeout:    tea.Int(timeoutMs),
		ConnectTimeout: tea.Int(timeoutMs),
	})
	if err != nil {
		status, code := sdkFault(err)
		return status, code, err
	}
	return http.StatusOK, "", nil
}

func sdkFault(err error) (int, string) {
	var teaErr *tea.SDKError
	if errors.As(err, &teaErr) && teaErr != nil {
		return tea.IntValue(teaErr.StatusCode), tea.StringValue(teaErr.Code)
	}
	var daraErr *dara.SDKError
	if errors.As(err, &daraErr) && daraErr != nil {
		return dara.IntValue(daraErr.StatusCode), dara.StringValue(daraErr.Code)
	}
	return 0, ""
}
