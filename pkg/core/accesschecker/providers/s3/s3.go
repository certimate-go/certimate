package s3

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	miniogo "github.com/minio/minio-go/v7"

	"github.com/certimate-go/certimate/internal/domain"
	s3tool "github.com/certimate-go/certimate/internal/tools/s3"
	"github.com/certimate-go/certimate/pkg/core"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
	CheckItem   = core.AccessCheckItem
)

type checker struct{}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeS3)
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

	diagnostics := []string{endpointHost(cfg.Endpoint)}
	client, err := s3tool.NewClient(&s3tool.Config{
		Endpoint:         cfg.Endpoint,
		AccessKey:        cfg.AccessKey,
		SecretKey:        cfg.SecretKey,
		SignatureVersion: cfg.SignatureVersion,
		UsePathStyle:     cfg.UsePathStyle,
		Region:           "",
		SkipTlsVerify:    cfg.AllowInsecureConnections,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: s3 client config is invalid", core.ErrAccessCheckConfigInvalid)
	}

	_, err = client.RawClient().ListBuckets(ctx)
	if err == nil {
		return passedResult(append(diagnostics, "credential accepted; permissions on specific resources were not checked")), nil
	}
	if errors.Is(err, context.Canceled) {
		return inconclusiveResult(core.AccessCheckReasonCanceled, core.AccessCheckItemStatusSkipped, diagnostics), nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return inconclusiveResult(core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, diagnostics), nil
	}
	if isTlsFailure(err) {
		return connectFailedResult(core.AccessCheckReasonTlsFailed, append(diagnostics, "tls handshake failed")), nil
	}

	resp := miniogo.ToErrorResponse(err)
	if resp.Code != "" || resp.StatusCode != 0 {
		diag := diagnostics
		if resp.StatusCode != 0 {
			diag = append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", resp.StatusCode))
		}
		if s3AuthRejected(resp) {
			return authRejectedResult(diag), nil
		}
		return inconclusiveResult(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, diag), nil
	}

	return connectFailedResult(core.AccessCheckReasonConnectFailed, append(diagnostics, "tcp or dns failure")), nil
}

type checkConfig struct {
	Endpoint                 string
	AccessKey                string
	SecretKey                string
	SignatureVersion         string
	UsePathStyle             bool
	AllowInsecureConnections bool
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForS3
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		Endpoint:                 strings.TrimSpace(wire.Endpoint),
		AccessKey:                strings.TrimSpace(wire.AccessKey),
		SecretKey:                wire.SecretKey,
		SignatureVersion:         strings.ToLower(strings.TrimSpace(wire.SignatureVersion)),
		UsePathStyle:             wire.UsePathStyle,
		AllowInsecureConnections: wire.AllowInsecureConnections,
	}
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("%w: endpoint is required", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.AccessKey == "" {
		return nil, fmt.Errorf("%w: accessKey is required", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.SecretKey == "" {
		return nil, fmt.Errorf("%w: secretKey is required", core.ErrAccessCheckConfigInvalid)
	}
	switch cfg.SignatureVersion {
	case "", s3tool.SignatureV2, s3tool.SignatureV4:
	default:
		return nil, fmt.Errorf("%w: signatureVersion is invalid", core.ErrAccessCheckConfigInvalid)
	}

	return cfg, nil
}

func s3AuthRejected(resp miniogo.ErrorResponse) bool {
	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return true
	}
	switch resp.Code {
	case miniogo.AccessDenied, miniogo.InvalidAccessKeyID, miniogo.SignatureDoesNotMatch, miniogo.AuthorizationHeaderMalformed,
		"ExpiredToken", "InvalidToken", "AllAccessDisabled":
		return true
	default:
		return false
	}
}

func endpointHost(endpoint string) string {
	if u, err := url.Parse(endpoint); err == nil && u.Host != "" {
		return u.Host
	}
	trimmed := strings.TrimPrefix(endpoint, "//")
	if i := strings.Index(trimmed, "/"); i >= 0 {
		trimmed = trimmed[:i]
	}
	return trimmed
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
