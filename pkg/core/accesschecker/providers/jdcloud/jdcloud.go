package jdcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	jdccore "github.com/go-acme/jdcloud-sdk-go/core"
	"github.com/go-acme/jdcloud-sdk-go/services/domainservice/apis"
	jdcclient "github.com/go-acme/jdcloud-sdk-go/services/domainservice/client"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/accesschecker"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
)

const (
	apiEndpoint = "domainservice.jdcloud-api.com"
	apiScheme   = "HTTPS"
	apiRegion   = "cn-north-1"
)

type checker struct {
	endpoint string
	scheme   string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{endpoint: apiEndpoint, scheme: apiScheme}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeJDCloud)
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
		endpoint = apiEndpoint
	}
	if scheme == "" {
		scheme = apiScheme
	}

	diagnostics := []string{endpoint}
	status, code, notes, callErr := describeDomains(ctx, endpoint, scheme, cfg.AccessKeyId, cfg.AccessKeySecret)
	diagnostics = append(diagnostics, notes...)
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

	var wire domain.AccessConfigForJDCloud
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

func describeDomains(ctx context.Context, endpoint, scheme, accessKeyId, accessKeySecret string) (int, string, []string, error) {
	credential := jdccore.NewCredentials(accessKeyId, accessKeySecret)
	client := jdcclient.NewDomainserviceClient(credential)
	client.DisableLogger()
	sdkConfig := jdccore.NewConfig()
	sdkConfig.SetScheme(scheme)
	sdkConfig.SetEndpoint(endpoint)
	sdkConfig.SetTimeout(accesschecker.BudgetLeft(ctx))
	client.SetConfig(sdkConfig)

	request := apis.NewDescribeDomainsRequestWithoutParam()
	request.SetRegionId(apiRegion)
	request.SetPageNumber(1)
	request.SetPageSize(1)

	status := http.StatusOK
	code := ""
	notes := []string{}
	callErr := accesschecker.Call(ctx, func() error {
		response, err := jdcclient.DescribeDomains(client, request)
		if err != nil {
			return err
		}
		if response.Error.Code != 0 || response.Error.Status != "" || response.Error.Message != "" {
			status, code = errorFault(response.Error)
			if response.Error.Message != "" {
				notes = append(notes, response.Error.Message)
			}
		}
		return nil
	})
	return status, code, notes, callErr
}

func errorFault(fault jdccore.ErrorResponse) (int, string) {
	detail := strings.ToLower(fault.Status + " " + fault.Message)
	if strings.Contains(detail, "signature") || strings.Contains(detail, "accesskey") || strings.Contains(detail, "credential") {
		return http.StatusUnauthorized, fault.Status
	}
	if fault.Code >= 400 && fault.Code < 600 {
		return fault.Code, fault.Status
	}
	return 0, fault.Status
}
