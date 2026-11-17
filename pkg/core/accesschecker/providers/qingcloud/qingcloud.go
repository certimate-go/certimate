package qingcloud

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/accesschecker"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
)

const apiEndpoint = "http://api.routewize.com"

type checker struct {
	endpoint string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{endpoint: apiEndpoint}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeQingCloud)
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
	status, code, notes, callErr := describeUserZones(ctx, endpoint, cfg.AccessKeyId, cfg.SecretAccessKey)
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

	var wire domain.AccessConfigForQingCloud
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
	Code    *int    `json:"code,omitempty"`
	Message *string `json:"message,omitempty"`
}

func (r *probeResponse) message() string {
	if r.Message == nil {
		return ""
	}
	return *r.Message
}

func describeUserZones(ctx context.Context, endpoint, accessKeyId, secretAccessKey string) (int, string, []string, error) {
	target := strings.TrimSuffix(endpoint, "/") + "/v1/user/zones"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return 0, "", nil, err
	}
	values := url.Values{}
	values.Set("limit", "1")
	values.Set("offset", "0")
	req.URL.RawQuery = values.Encode()

	date := time.Now().UTC().Format(time.RFC1123)
	canonicalResource := req.URL.Path + "?" + req.URL.Query().Encode()
	mac := hmac.New(sha256.New, []byte(secretAccessKey))
	mac.Write([]byte(req.Method + "\n" + date + "\n" + canonicalResource))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	req.Header.Set("Date", date)
	req.Header.Set("Authorization", fmt.Sprintf("QC-HMAC-SHA256 %s:%s", accessKeyId, signature))

	client := &http.Client{Timeout: accesschecker.BudgetLeft(ctx)}
	resp, callErr := client.Do(req)
	if callErr != nil {
		return 0, "", nil, callErr
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return 0, "", nil, err
	}

	res := &probeResponse{}
	_ = json.Unmarshal(body, res)

	bodyCode := ""
	notes := []string{}
	if res.Code != nil && *res.Code != 0 {
		bodyCode = strconv.Itoa(*res.Code)
		if res.Message != nil && *res.Message != "" {
			notes = append(notes, *res.Message)
		}
	}

	if resp.StatusCode == http.StatusOK && bodyCode == "" {
		return http.StatusOK, "", nil, nil
	}

	if credentialFault(bodyCode + " " + res.message()) {
		return http.StatusUnauthorized, bodyCode, notes, nil
	}
	if bodyCode != "" && resp.StatusCode == http.StatusOK {
		return 0, bodyCode, notes, nil
	}
	return resp.StatusCode, bodyCode, notes, nil
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
