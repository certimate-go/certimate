package volcengine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/volcengine/volc-sdk-golang/base"
	dns "github.com/volcengine/volc-sdk-golang/service/dns"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/accesschecker"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
)

const (
	apiHost   = "open.volcengineapi.com"
	apiScheme = "https"
	apiRegion = "cn-beijing"
)

type checker struct {
	host   string
	scheme string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{host: apiHost, scheme: apiScheme}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeVolcEngine)
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

	host := c.host
	scheme := c.scheme
	if host == "" {
		host = apiHost
	}
	if scheme == "" {
		scheme = apiScheme
	}

	diagnostics := []string{host}
	status, code, callErr := listZones(ctx, host, scheme, cfg.AccessKeyId, cfg.SecretAccessKey)
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

	var wire domain.AccessConfigForVolcEngine
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

func listZones(ctx context.Context, host, scheme, accessKeyId, secretAccessKey string) (int, string, error) {
	client := newDNSClient(host, scheme, accessKeyId, secretAccessKey, accesschecker.BudgetLeft(ctx))
	pageSize := "1"
	pageNumber := "1"
	_, err := client.ListZones(ctx, &dns.ListZonesRequest{PageSize: &pageSize, PageNumber: &pageNumber})
	if err != nil {
		return topFault(err)
	}
	return http.StatusOK, "", nil
}

func newDNSClient(host, scheme, accessKeyId, secretAccessKey string, timeout time.Duration) *dns.Client {
	info := &base.ServiceInfo{
		Timeout: timeout,
		Host:    host,
		Header:  http.Header{"Accept": []string{"application/json"}},
		Scheme:  scheme,
		Credentials: base.Credentials{
			Service:         dns.ServiceName,
			Region:          apiRegion,
			AccessKeyID:     accessKeyId,
			SecretAccessKey: secretAccessKey,
		},
	}
	raw := base.NewClient(info, nil)
	caller := &dns.VolcCaller{Volc: raw}
	caller.Volc.SetAccessKey(accessKeyId)
	caller.Volc.SetSecretKey(secretAccessKey)
	caller.Volc.SetHost(host)
	caller.Volc.SetScheme(scheme)
	caller.Volc.SetTimeout(timeout)
	return dns.NewClient(caller)
}

func topFault(err error) (int, string, error) {
	var top *dns.TOPError
	if errors.As(err, &top) && top != nil {
		return 0, top.Code, err
	}
	return 0, "", err
}
