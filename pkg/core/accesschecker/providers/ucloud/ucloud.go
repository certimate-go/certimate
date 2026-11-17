package ucloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/accesschecker"
	"github.com/certimate-go/certimate/pkg/sdk3rd/ucloud/udnr"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
)

const apiEndpoint = "https://api.ucloud.cn"

const authFailureRetCode = 230

type checker struct {
	endpoint string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{endpoint: apiEndpoint}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeUCloud)
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
	status, code, notes, callErr := getProjectList(ctx, endpoint, cfg)
	diagnostics = append(diagnostics, notes...)
	return accesschecker.Finish(callErr, status, code, diagnostics, cfg.PublicKey, cfg.PrivateKey), nil
}

type checkConfig struct {
	PublicKey  string
	PrivateKey string
	ProjectId  string
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForUCloud
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		PublicKey:  strings.TrimSpace(wire.PublicKey),
		PrivateKey: wire.PrivateKey,
		ProjectId:  strings.TrimSpace(wire.ProjectId),
	}
	if cfg.PublicKey == "" {
		return nil, fmt.Errorf("%w: publicKey is required", core.ErrAccessCheckConfigInvalid)
	}
	if strings.TrimSpace(cfg.PrivateKey) == "" {
		return nil, fmt.Errorf("%w: privateKey is required", core.ErrAccessCheckConfigInvalid)
	}
	return cfg, nil
}

type probeRequest struct {
	request.CommonBase
}

type probeResponse struct {
	response.CommonBase
}

type ucloudFault interface {
	error
	Name() string
	StatusCode() int
	Code() int
	Message() string
	OriginError() error
}

func getProjectList(ctx context.Context, endpoint string, cfg *checkConfig) (int, string, []string, error) {
	sdkConfig := ucloud.NewConfig()
	sdkConfig.BaseUrl = endpoint
	sdkConfig.ProjectId = cfg.ProjectId

	credential := auth.NewCredential()
	credential.PublicKey = cfg.PublicKey
	credential.PrivateKey = cfg.PrivateKey

	client := udnr.NewClient(&sdkConfig, &credential)

	req := &probeRequest{}
	client.SetupRequest(req)
	req.WithTimeout(accesschecker.BudgetLeft(ctx))

	res := &probeResponse{}
	callErr := accesschecker.Call(ctx, func() error {
		return client.InvokeAction("GetProjectList", req, res)
	})
	if callErr != nil {
		var fault ucloudFault
		if errors.As(callErr, &fault) {
			status, code, notes := faultFault(fault)
			origin := fault.OriginError()
			if origin == nil {
				origin = callErr
			}
			return status, code, notes, origin
		}
		return 0, "", nil, callErr
	}
	return http.StatusOK, "", nil, nil
}

func faultFault(fault ucloudFault) (int, string, []string) {
	notes := []string{}
	if message := fault.Message(); message != "" {
		notes = append(notes, message)
	}
	if fault.Code() > 0 {
		code := strconv.Itoa(fault.Code())
		if fault.Code() == authFailureRetCode || credentialFault(fault.Message()) {
			return http.StatusUnauthorized, code, notes
		}
		return 0, code, notes
	}
	if statusCode := fault.StatusCode(); statusCode >= 400 && statusCode < 600 {
		return statusCode, "", notes
	}
	return 0, fault.Name(), notes
}

func credentialFault(detail string) bool {
	detail = strings.ToLower(detail)
	for _, keyword := range []string{"signature", "public key", "private key", "unauthorized", "authentica", "credential"} {
		if strings.Contains(detail, keyword) {
			return true
		}
	}
	return false
}
