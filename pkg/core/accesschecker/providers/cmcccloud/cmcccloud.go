package cmcccloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/accesschecker"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
)

const (
	apiEndpoint = "https://ecloud.10086.cn/api/query/openapi/apim/request/sdk"
	probeAction = "deleteRecordOpenapi"
	probeNote   = "probe sends a no-op empty record-list delete; credentials count as verified only when the server reports state OK"
)

var authTextMarkers = []string{"accesskey", "access key", "secretkey", "secret key", "signature", "credential", "unauthorized"}

type checker struct {
	endpoint string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{endpoint: apiEndpoint}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeCMCCCloud)
}

func (c *checker) Scope() []string {
	return []string{core.AccessCheckItemConnect, core.AccessCheckItemAuth}
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

	var wire domain.AccessConfigForCMCCCloud
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

type cloudRequest struct {
	Product       string `json:"product"`
	Version       string `json:"version"`
	SDKVersion    string `json:"sdkVersion"`
	Language      string `json:"language"`
	API           string `json:"api"`
	PoolID        string `json:"poolId"`
	AccessKey     string `json:"accessKey"`
	SecretKey     string `json:"secretKey"`
	BodyParameter any    `json:"bodyParameter"`
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
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, fmt.Errorf("%w: probe endpoint is invalid", core.ErrAccessCheckConfigInvalid)
	}

	body, err := json.Marshal(cloudRequest{
		Product:       "clouddns",
		Version:       "v1",
		SDKVersion:    "1.0.1",
		Language:      "Golang",
		API:           probeAction,
		PoolID:        "CIDC-CORE-00",
		AccessKey:     cfg.AccessKeyId,
		SecretKey:     cfg.AccessKeySecret,
		BodyParameter: map[string]any{"recordIdList": []string{}},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: request body is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: probe endpoint is invalid", core.ErrAccessCheckConfigInvalid)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	diagnostics := []string{u.Host}
	secrets := []string{cfg.AccessKeyId, cfg.AccessKeySecret}
	client := accesschecker.HTTPClient()
	client.Timeout = accesschecker.BudgetLeft(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return accesschecker.Finish(err, 0, "", diagnostics, secrets...), nil
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", resp.StatusCode), "response body unreadable", probeNote)
		return accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, diag), secrets...), nil
	}

	return classify(resp.StatusCode, raw, diagnostics, secrets), nil
}

func classify(status int, raw []byte, diagnostics []string, secrets []string) *CheckResult {
	if status >= http.StatusBadRequest {
		return accesschecker.Finish(nil, status, "", diagnostics, secrets...)
	}

	var gateway struct {
		Body struct {
			ResponseBody string `json:"responseBody"`
		} `json:"body"`
	}
	if err := json.Unmarshal(raw, &gateway); err != nil || gateway.Body.ResponseBody == "" {
		diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", status), "gateway response unreadable", probeNote)
		return accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, diag), secrets...)
	}

	var result struct {
		State        string `json:"state"`
		ErrorCode    string `json:"errorCode"`
		ErrorMessage string `json:"errorMessage"`
	}
	if err := json.Unmarshal([]byte(gateway.Body.ResponseBody), &result); err != nil {
		diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", status), "gateway response unreadable", probeNote)
		return accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, diag), secrets...)
	}

	diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", status))
	if result.ErrorCode != "" {
		diag = append(diag, "code "+result.ErrorCode)
	}
	if result.ErrorMessage != "" {
		diag = append(diag, result.ErrorMessage)
	}

	if result.State == "OK" {
		return accesschecker.Redact(accesschecker.Passed(append(diag, probeNote)), secrets...)
	}
	if looksLikeAuthFailure(result.ErrorCode, result.ErrorMessage) {
		return accesschecker.Redact(accesschecker.AuthRejected(append(diag, probeNote)), secrets...)
	}
	return accesschecker.Redact(accesschecker.Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, append(diag, "no credential verdict", probeNote)), secrets...)
}

func looksLikeAuthFailure(code, message string) bool {
	if authCodeText(code) {
		return true
	}
	msg := strings.ToLower(message)
	for _, marker := range authTextMarkers {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}

func authCodeText(code string) bool {
	c := strings.ToLower(strings.TrimSpace(code))
	return strings.HasPrefix(c, "invalidaccesskey") ||
		strings.HasPrefix(c, "signaturedoesnotmatch") ||
		strings.HasPrefix(c, "authfailure") ||
		strings.HasPrefix(c, "invalidcredential") ||
		strings.HasPrefix(c, "unauthorized")
}
