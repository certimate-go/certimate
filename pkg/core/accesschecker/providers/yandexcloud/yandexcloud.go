package yandexcloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/go-sdk/v2/credentials"
	"github.com/yandex-cloud/go-sdk/v2/pkg/endpoints"
	"github.com/yandex-cloud/go-sdk/v2/pkg/iamkey"
	"github.com/yandex-cloud/go-sdk/v2/pkg/options"

	dnsv1 "github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
	dnssdk "github.com/yandex-cloud/go-sdk/services/dns/v1"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/accesschecker"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
)

const apiEndpoint = "api.cloud.yandex.net:443"

type checker struct {
	resolverEndpoint string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeYandexCloud)
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

	key, err := iamkey.ReadFromJSONBytes([]byte(cfg.ServiceAccountKey))
	if err != nil {
		return nil, fmt.Errorf("%w: serviceAccountKey is malformed", core.ErrAccessCheckConfigInvalid)
	}
	creds, err := credentials.ServiceAccountKey(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", core.ErrAccessCheckConfigInvalid, err)
	}

	sdkOpts := []options.Option{options.WithCredentials(creds)}
	diagnostics := []string{apiEndpoint}
	if c.resolverEndpoint != "" {
		diagnostics = []string{c.resolverEndpoint}
		sdkOpts = append(sdkOpts,
			options.WithPlaintext(),
			options.WithEndpointsResolver(endpoints.NewSingleEndpointResolver(c.resolverEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))),
		)
	}

	status, code, callErr := listDnsZones(ctx, cfg.FolderId, sdkOpts)
	return accesschecker.Finish(callErr, status, code, diagnostics, cfg.ServiceAccountKey), nil
}

type checkConfig struct {
	FolderId          string
	ServiceAccountKey string
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForYandexCloud
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		FolderId:          strings.TrimSpace(wire.FolderId),
		ServiceAccountKey: wire.ServiceAccountKey,
	}
	if cfg.FolderId == "" {
		return nil, fmt.Errorf("%w: folderId is required", core.ErrAccessCheckConfigInvalid)
	}
	if strings.TrimSpace(cfg.ServiceAccountKey) == "" {
		return nil, fmt.Errorf("%w: serviceAccountKey is required", core.ErrAccessCheckConfigInvalid)
	}
	return cfg, nil
}

type grpcStatus interface {
	GRPCStatus() *status.Status
}

func listDnsZones(ctx context.Context, folderId string, sdkOpts []options.Option) (int, string, error) {
	sdk, err := ycsdk.Build(ctx, sdkOpts...)
	if err != nil {
		return grpcFault(err)
	}

	client := dnssdk.NewDnsZoneClient(sdk)
	_, err = client.List(ctx, &dnsv1.ListDnsZonesRequest{
		FolderId: folderId,
		PageSize: 1,
	})
	if err != nil {
		return grpcFault(err)
	}
	return http.StatusOK, "", nil
}

func grpcFault(err error) (int, string, error) {
	var st grpcStatus
	if errors.As(err, &st) {
		mapped, mappedErr, transportMessage := grpcFaultError(st)
		if mappedErr != nil {
			return 0, "", mappedErr
		}
		if transportMessage != "" {
			return 0, "", errors.New(transportMessage)
		}
		if mapped != 0 {
			return mapped, "", nil
		}
		return 0, st.GRPCStatus().Code().String(), nil
	}
	return 0, "", err
}

func grpcFaultError(st grpcStatus) (int, error, string) {
	switch st.GRPCStatus().Code() {
	case codes.Unauthenticated:
		if transportUndercurrent(st.GRPCStatus().Message()) {
			return 0, nil, st.GRPCStatus().Message()
		}
		return http.StatusUnauthorized, nil, ""
	case codes.PermissionDenied:
		return http.StatusForbidden, nil, ""
	case codes.DeadlineExceeded:
		return 0, context.DeadlineExceeded, ""
	case codes.Canceled:
		return 0, context.Canceled, ""
	default:
		return 0, nil, ""
	}
}

func transportUndercurrent(message string) bool {
	for _, marker := range []string{"connection refused", "dial tcp", "no such host", "network is unreachable", "context deadline exceeded", "Client.Timeout exceeded", "tls: ", "x509:"} {
		if strings.Contains(message, marker) {
			return true
		}
	}
	return false
}
