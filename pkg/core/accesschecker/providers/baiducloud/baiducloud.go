package baiducloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/baidubce/bce-sdk-go/bce"
	baidudns "github.com/baidubce/bce-sdk-go/services/dns"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/accesschecker"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
)

const apiEndpoint = "https://dns.baidubce.com"

type checker struct {
	endpoint string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{endpoint: apiEndpoint}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeBaiduCloud)
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
	if endpoint == "" {
		endpoint = apiEndpoint
	}
	diagnostics := []string{accesschecker.HostOf(endpoint)}
	status, code, callErr := listZones(ctx, endpoint, cfg.AccessKeyId, cfg.SecretAccessKey)
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

	var wire domain.AccessConfigForBaiduCloud
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

func listZones(ctx context.Context, endpoint, accessKeyId, secretAccessKey string) (int, string, error) {
	client, err := baidudns.NewClient(accessKeyId, secretAccessKey, endpoint)
	if err != nil {
		return 0, "", err
	}
	millis := int(accesschecker.BudgetLeft(ctx) / time.Millisecond)
	if millis < 1000 {
		millis = 1000
	}
	client.Config.ConnectionTimeoutInMillis = millis

	var callErr error
	err = accesschecker.Call(ctx, func() error {
		_, callErr = client.ListZone(&baidudns.ListZoneRequest{MaxKeys: 1})
		return callErr
	})
	if err != nil {
		return sdkFault(err)
	}
	return 200, "", nil
}

func sdkFault(err error) (int, string, error) {
	var svcErr *bce.BceServiceError
	if errors.As(err, &svcErr) && svcErr != nil {
		return svcErr.StatusCode, svcErr.Code, err
	}
	return 0, "", err
}
