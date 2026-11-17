package ctcccloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/accesschecker"
	common "github.com/certimate-go/certimate/pkg/sdk3rd/ctyun/zz-shared-common"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
)

const (
	apiEndpoint = "https://smartdns-global.ctapi.ctyun.cn"
	probeDomain = "example.com"
)

type checker struct {
	endpoint string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{endpoint: apiEndpoint}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeCTCCCloud)
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
	status, code, notes, callErr := queryRecordList(ctx, endpoint, cfg.AccessKeyId, cfg.SecretAccessKey)
	diagnostics = append(diagnostics, notes...)
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

	var wire domain.AccessConfigForCTCCCloud
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

type probeResponse struct {
	StatusCode   json.RawMessage `json:"statusCode,omitempty"`
	Message      *string         `json:"message,omitempty"`
	Error        *string         `json:"error,omitempty"`
	ErrorMessage *string         `json:"errorMessage,omitempty"`
}

func (r *probeResponse) bodyCode() string {
	if len(r.StatusCode) == 0 {
		return ""
	}
	var asString string
	if err := json.Unmarshal(r.StatusCode, &asString); err == nil {
		return asString
	}
	var asNumber json.Number
	if err := json.Unmarshal(r.StatusCode, &asNumber); err == nil {
		return asNumber.String()
	}
	return ""
}

func (r *probeResponse) bodyFault() string {
	faults := []string{}
	if r.Error != nil && *r.Error != "" {
		faults = append(faults, *r.Error)
	}
	if r.ErrorMessage != nil && *r.ErrorMessage != "" {
		faults = append(faults, *r.ErrorMessage)
	}
	if r.Message != nil && *r.Message != "" {
		faults = append(faults, *r.Message)
	}
	return strings.Join(faults, "; ")
}

func (r *probeResponse) bodyAccepted() bool {
	if code := r.bodyCode(); code != "" && code != "200" {
		return false
	}
	return r.Error == nil || *r.Error == ""
}

func queryRecordList(ctx context.Context, endpoint, accessKeyId, secretAccessKey string) (int, string, []string, error) {
	client, err := common.NewClient(endpoint, common.WithAkSk(accessKeyId, secretAccessKey))
	if err != nil {
		return 0, "", nil, fmt.Errorf("%w: %v", core.ErrAccessCheckConfigInvalid, err)
	}
	client.SetTimeout(accesschecker.BudgetLeft(ctx))

	req, err := client.NewRequest(http.MethodGet, "/v2/queryRecordList")
	if err != nil {
		return 0, "", nil, err
	}
	req.SetContext(ctx)
	req.SetQueryParam("domain", probeDomain)

	res := &probeResponse{}
	resp, callErr := client.DoRequestWithResult(req, res)
	if resp == nil || resp.StatusCode() == 0 {
		return 0, "", nil, callErr
	}

	if resp.StatusCode() == http.StatusOK && res.bodyAccepted() {
		return http.StatusOK, "", nil, nil
	}

	notes := []string{}
	if detail := res.bodyFault(); detail != "" {
		notes = append(notes, detail)
	}
	bodyCode := res.bodyCode()
	if credentialFault(bodyCode + " " + res.bodyFault()) {
		return http.StatusUnauthorized, bodyCode, notes, nil
	}
	return resp.StatusCode(), bodyCode, notes, nil
}

func credentialFault(detail string) bool {
	detail = strings.ToLower(detail)
	for _, keyword := range []string{"signature", "accesskey", "access key", "unauthorized", "authentica", "credential", "secret"} {
		if strings.Contains(detail, keyword) {
			return true
		}
	}
	return false
}
