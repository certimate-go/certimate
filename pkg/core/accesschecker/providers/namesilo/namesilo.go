package namesilo

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
	CheckItem   = core.AccessCheckItem
)

const apiBaseURL = "https://www.namesilo.com/api"

const (
	namesiloCodeSuccess      = "300"
	namesiloCodeRegistration = "301"
	namesiloCodeOrder        = "302"
	namesiloCodeInvalidKey   = "110"
)

type checker struct {
	baseURL string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{baseURL: apiBaseURL}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeNameSilo)
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

	secret, err := parseConfig(config)
	if err != nil {
		return nil, err
	}

	endpoint, err := probeEndpoint(c.baseURL, secret)
	if err != nil {
		return nil, fmt.Errorf("%w: probe url is invalid", core.ErrAccessCheckConfigInvalid)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: probe url is invalid", core.ErrAccessCheckConfigInvalid)
	}
	req.Header.Set("Accept", "application/xml")

	diagnostics := []string{endpoint.Host}
	resp, err := newHTTPClient().Do(req)
	if err != nil {
		return transportResult(err, diagnostics), nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", resp.StatusCode), "response body unreadable")
		return inconclusiveResult(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, diag), nil
	}

	return classifyNamesilo(resp.StatusCode, body, diagnostics), nil
}

func parseConfig(config map[string]any) (string, error) {
	if len(config) == 0 {
		return "", fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForNameSilo
	if err := json.Unmarshal(data, &wire); err != nil {
		return "", fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	secret := strings.TrimSpace(wire.ApiKey)
	if secret == "" {
		return "", fmt.Errorf("%w: apiKey is required", core.ErrAccessCheckConfigInvalid)
	}

	return secret, nil
}

func probeEndpoint(base, apiKey string) (*url.URL, error) {
	if strings.TrimSpace(base) == "" {
		base = apiBaseURL
	}
	endpoint, err := url.Parse(base)
	if err != nil || endpoint.Host == "" || (endpoint.Scheme != "http" && endpoint.Scheme != "https") {
		return nil, fmt.Errorf("probe url is invalid")
	}
	endpoint = endpoint.JoinPath("listDomains")
	if !strings.HasPrefix(endpoint.Path, "/") {
		endpoint.Path = "/" + endpoint.Path
	}
	query := endpoint.Query()
	query.Set("version", "1")
	query.Set("type", "xml")
	query.Set("key", apiKey)
	endpoint.RawQuery = query.Encode()
	return endpoint, nil
}

type namesiloResponse struct {
	Reply struct {
		Code string `xml:"code"`
	} `xml:"reply"`
}

func classifyNamesilo(status int, body []byte, diagnostics []string) *CheckResult {
	diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", status))
	if status == http.StatusUnauthorized {
		return authRejectedResult(diag)
	}
	if status == http.StatusForbidden {
		return inconclusiveResult(core.AccessCheckReasonProbeForbidden, core.AccessCheckItemStatusPassed, diag)
	}
	if status < 200 || status >= 300 {
		return inconclusiveResult(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, diag)
	}

	var payload namesiloResponse
	if err := xml.Unmarshal(body, &payload); err != nil {
		return inconclusiveResult(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, append(diag, "no credential verdict"))
	}

	switch strings.TrimSpace(payload.Reply.Code) {
	case namesiloCodeSuccess, namesiloCodeRegistration, namesiloCodeOrder:
		return passedResult(append(diag, "reply code "+strings.TrimSpace(payload.Reply.Code), "credential accepted; permissions on specific resources were not checked"))
	case namesiloCodeInvalidKey:
		return authRejectedResult(append(diag, "reply code "+namesiloCodeInvalidKey))
	default:
		code := strings.TrimSpace(payload.Reply.Code)
		if code == "" {
			code = "empty"
		}
		return inconclusiveResult(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, append(diag, "reply code "+code))
	}
}

func newHTTPClient() *http.Client {
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func classifyStatus(status int, diagnostics []string) *CheckResult {
	diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", status))
	switch {
	case status >= 200 && status < 300:
		return passedResult(append(diag, "credential accepted; permissions on specific resources were not checked"))
	case status == http.StatusUnauthorized:
		return authRejectedResult(diag)
	case status == http.StatusForbidden:
		return inconclusiveResult(core.AccessCheckReasonProbeForbidden, core.AccessCheckItemStatusPassed, diag)
	default:
		return inconclusiveResult(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, diag)
	}
}

func transportResult(err error, diagnostics []string) *CheckResult {
	switch {
	case errors.Is(err, context.Canceled):
		return inconclusiveResult(core.AccessCheckReasonCanceled, core.AccessCheckItemStatusSkipped, diagnostics)
	case errors.Is(err, context.DeadlineExceeded):
		return inconclusiveResult(core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, diagnostics)
	case isTlsFailure(err):
		return connectFailedResult(core.AccessCheckReasonTlsFailed, append(append([]string{}, diagnostics...), "tls handshake failed"))
	default:
		return connectFailedResult(core.AccessCheckReasonConnectFailed, append(append([]string{}, diagnostics...), "tcp or dns failure"))
	}
}

func isTlsFailure(err error) bool {
	var certVerifyErr *tls.CertificateVerificationError
	if errors.As(err, &certVerifyErr) {
		return true
	}
	var unknownAuth x509.UnknownAuthorityError
	if errors.As(err, &unknownAuth) {
		return true
	}
	var hostnameErr x509.HostnameError
	if errors.As(err, &hostnameErr) {
		return true
	}
	var alertErr tls.AlertError
	if errors.As(err, &alertErr) {
		return true
	}
	var recordHeaderErr *tls.RecordHeaderError
	return errors.As(err, &recordHeaderErr)
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
