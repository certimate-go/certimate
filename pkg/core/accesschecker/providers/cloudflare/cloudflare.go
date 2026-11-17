package cloudflare

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
	CheckItem   = core.AccessCheckItem
)

const apiBaseUrl = "https://api.cloudflare.com/client/v4"

const (
	tokenStatusActive   = "active"
	tokenStatusDisabled = "disabled"
	tokenStatusExpired  = "expired"
	tokenStatusRevoked  = "revoked"
)

const (
	tokenKindMain = "apiToken"
	tokenKindZone = "apiTokenForZone"
)

const (
	verdictPassed int = iota
	verdictRejected
	verdictForbidden
	verdictRemoteError
)

type checker struct {
	baseURL string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{baseURL: apiBaseUrl}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeCloudflare)
}

func (c *checker) Scope() []string {
	return []string{core.AccessCheckItemConnect, core.AccessCheckItemAuth}
}

func (c *checker) Check(ctx context.Context, config map[string]any) (*CheckResult, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, core.AccessCheckDefaultBudget)
		defer cancel()
	}

	cfg, err := parseConfig(config)
	if err != nil {
		return nil, err
	}

	base := lo.Ternary(c.baseURL != "", c.baseURL, apiBaseUrl)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	diagnostics := []string{hostOf(base)}
	tokens := []token{{kind: tokenKindMain, value: cfg.ApiToken}}
	if cfg.ApiTokenForZone != "" {
		tokens = append(tokens, token{kind: tokenKindZone, value: cfg.ApiTokenForZone})
	}

	verdicts := make([]verdict, 0, len(tokens))
	for _, tok := range tokens {
		v, err := c.verifyToken(ctx, client, base, tok)
		if err != nil {
			return transportResult(err, diagnostics), nil
		}
		diagnostics = append(diagnostics, v.diagnostic)
		verdicts = append(verdicts, v)
	}

	return combineVerdicts(verdicts, diagnostics), nil
}

type token struct {
	kind  string
	value string
}

type verdict struct {
	kind       string
	outcome    int
	diagnostic string
}

type configWire struct {
	ApiToken        string `json:"apiToken"`
	ApiTokenForZone string `json:"apiTokenForZone,omitempty"`
}

type checkConfig struct {
	ApiToken        string
	ApiTokenForZone string
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire configWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		ApiToken:        strings.TrimSpace(wire.ApiToken),
		ApiTokenForZone: strings.TrimSpace(wire.ApiTokenForZone),
	}
	if cfg.ApiToken == "" {
		return nil, fmt.Errorf("%w: apiToken is required", core.ErrAccessCheckConfigInvalid)
	}

	return cfg, nil
}

func (c *checker) verifyToken(ctx context.Context, client *http.Client, base string, tok token) (verdict, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/user/tokens/verify", nil)
	if err != nil {
		return verdict{}, err
	}
	req.Header.Set("Authorization", "Bearer "+tok.value)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return verdict{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return verdict{
			kind:       tok.kind,
			outcome:    verdictRemoteError,
			diagnostic: fmt.Sprintf("token %s: response body unreadable", tok.kind),
		}, nil
	}

	var payload verifyResponse
	outcome, detail := classifyVerify(resp.StatusCode, &payload, json.Unmarshal(body, &payload))
	return verdict{
		kind:       tok.kind,
		outcome:    outcome,
		diagnostic: fmt.Sprintf("token %s: %s", tok.kind, detail),
	}, nil
}

type verifyResponse struct {
	Success bool        `json:"success"`
	Errors  []apiError  `json:"errors"`
	Result  *tokenState `json:"result"`
}

type apiError struct {
	Code       int        `json:"code"`
	Message    string     `json:"message"`
	ErrorChain []apiError `json:"error_chain,omitempty"`
}

type tokenState struct {
	Status string `json:"status"`
}

func classifyVerify(httpStatus int, payload *verifyResponse, parseErr error) (int, string) {
	if httpStatus == http.StatusForbidden {
		return verdictForbidden, "endpoint returned HTTP 403 without a credential verdict"
	}

	if parseErr == nil {
		if payload.Result != nil {
			status := strings.ToLower(strings.TrimSpace(payload.Result.Status))
			if payload.Success {
				switch status {
				case tokenStatusActive:
					return verdictPassed, "token is active"
				case tokenStatusDisabled, tokenStatusExpired, tokenStatusRevoked:
					return verdictRejected, "token status is " + status
				}
			}
		}
		if !payload.Success && reportsInvalidToken(payload.Errors) {
			return verdictRejected, fmt.Sprintf("api reported the token as invalid (%s)", firstErrorSummary(payload.Errors))
		}
	}

	detail := fmt.Sprintf("no credential verdict from response (HTTP %d)", httpStatus)
	if summary := firstErrorSummary(payload.Errors); summary != "" {
		detail = fmt.Sprintf("no credential verdict from response (HTTP %d: %s)", httpStatus, summary)
	}
	return verdictRemoteError, detail
}

func reportsInvalidToken(errs []apiError) bool {
	codes := make([]int, 0, len(errs)*2)
	for _, err := range errs {
		codes = append(codes, err.Code)
		for _, chained := range err.ErrorChain {
			codes = append(codes, chained.Code)
		}
	}
	return lo.SomeBy(codes, func(code int) bool {
		return code == 1000 || code == 1003 || code == 6111
	})
}

func firstErrorSummary(errs []apiError) string {
	if len(errs) == 0 {
		return ""
	}
	first := errs[0]
	if len(first.ErrorChain) > 0 {
		chained := first.ErrorChain[0]
		return fmt.Sprintf("%d %s: %d %s", first.Code, first.Message, chained.Code, chained.Message)
	}
	return fmt.Sprintf("%d %s", first.Code, first.Message)
}

func transportResult(err error, diagnostics []string) *CheckResult {
	switch {
	case errors.Is(err, context.Canceled):
		return inconclusiveResult(core.AccessCheckReasonCanceled, core.AccessCheckItemStatusSkipped, diagnostics)
	case errors.Is(err, context.DeadlineExceeded):
		return inconclusiveResult(core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, diagnostics)
	case isTlsFailure(err):
		return connectFailedResult(core.AccessCheckReasonTlsFailed, append(diagnostics, "tls handshake failed"))
	default:
		return connectFailedResult(core.AccessCheckReasonConnectFailed, append(diagnostics, "tcp or dns failure"))
	}
}

func isTlsFailure(err error) bool {
	var certVerifyErr *tls.CertificateVerificationError
	if errors.As(err, &certVerifyErr) {
		return true
	}
	var alertErr tls.AlertError
	if errors.As(err, &alertErr) {
		return true
	}
	var recordHeaderErr *tls.RecordHeaderError
	return errors.As(err, &recordHeaderErr)
}

func combineVerdicts(verdicts []verdict, diagnostics []string) *CheckResult {
	if lo.ContainsBy(verdicts, func(v verdict) bool { return v.outcome == verdictRejected }) {
		return authRejectedResult(diagnostics)
	}
	if lo.ContainsBy(verdicts, func(v verdict) bool { return v.outcome == verdictForbidden }) {
		return inconclusiveResult(core.AccessCheckReasonProbeForbidden, core.AccessCheckItemStatusPassed, diagnostics)
	}
	if lo.EveryBy(verdicts, func(v verdict) bool { return v.outcome == verdictPassed }) {
		return passedResult(append(diagnostics, "verified token validity and active status only; permissions on specific resources were not checked"))
	}
	return inconclusiveResult(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, diagnostics)
}

func hostOf(base string) string {
	if u, err := url.Parse(base); err == nil && u.Host != "" {
		return u.Host
	}
	return base
}

func passedResult(diagnostics []string) *CheckResult {
	return &CheckResult{
		Status:     core.AccessCheckStatusPassed,
		ReasonCode: "",
		Items: []CheckItem{
			{Id: core.AccessCheckItemConnect, Status: core.AccessCheckItemStatusPassed},
			{Id: core.AccessCheckItemAuth, Status: core.AccessCheckItemStatusPassed},
		},
		Diagnostics: diagnostics,
	}
}

func authRejectedResult(diagnostics []string) *CheckResult {
	return &CheckResult{
		Status:     core.AccessCheckStatusFailed,
		ReasonCode: core.AccessCheckReasonAuthRejected,
		Items: []CheckItem{
			{Id: core.AccessCheckItemConnect, Status: core.AccessCheckItemStatusPassed},
			{Id: core.AccessCheckItemAuth, Status: core.AccessCheckItemStatusFailed},
		},
		Diagnostics: diagnostics,
	}
}

func inconclusiveResult(reason string, connectStatus string, diagnostics []string) *CheckResult {
	return &CheckResult{
		Status:     core.AccessCheckStatusInconclusive,
		ReasonCode: reason,
		Items: []CheckItem{
			{Id: core.AccessCheckItemConnect, Status: connectStatus},
			{Id: core.AccessCheckItemAuth, Status: core.AccessCheckItemStatusSkipped},
		},
		Diagnostics: diagnostics,
	}
}

func connectFailedResult(reason string, diagnostics []string) *CheckResult {
	return &CheckResult{
		Status:     core.AccessCheckStatusFailed,
		ReasonCode: reason,
		Items: []CheckItem{
			{Id: core.AccessCheckItemConnect, Status: core.AccessCheckItemStatusFailed},
			{Id: core.AccessCheckItemAuth, Status: core.AccessCheckItemStatusSkipped},
		},
		Diagnostics: diagnostics,
	}
}
