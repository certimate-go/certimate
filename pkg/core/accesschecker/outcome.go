package accesschecker

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/certimate-go/certimate/pkg/core"
)

const acceptedNote = "credential accepted; permissions on specific resources were not checked"

func WithBudget(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, core.AccessCheckDefaultBudget)
}

func BudgetLeft(ctx context.Context) time.Duration {
	if deadline, ok := ctx.Deadline(); ok {
		left := time.Until(deadline)
		if left < time.Millisecond {
			return time.Millisecond
		}
		return left
	}
	return core.AccessCheckDefaultBudget
}

func HTTPClient() *http.Client {
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func HostOf(raw string) string {
	if u, err := url.Parse(raw); err == nil && u.Host != "" {
		return u.Host
	}
	return raw
}

func Passed(diagnostics []string) *core.AccessCheckResult {
	return &core.AccessCheckResult{
		Status:     core.AccessCheckStatusPassed,
		ReasonCode: "",
		Items: []core.AccessCheckItem{
			{Id: core.AccessCheckItemConnect, Status: core.AccessCheckItemStatusPassed},
			{Id: core.AccessCheckItemAuth, Status: core.AccessCheckItemStatusPassed},
		},
		Diagnostics: append(append([]string{}, diagnostics...), acceptedNote),
	}
}

func AuthRejected(diagnostics []string) *core.AccessCheckResult {
	return &core.AccessCheckResult{
		Status:     core.AccessCheckStatusFailed,
		ReasonCode: core.AccessCheckReasonAuthRejected,
		Items: []core.AccessCheckItem{
			{Id: core.AccessCheckItemConnect, Status: core.AccessCheckItemStatusPassed},
			{Id: core.AccessCheckItemAuth, Status: core.AccessCheckItemStatusFailed},
		},
		Diagnostics: diagnostics,
	}
}

func Inconclusive(reason string, connectStatus string, diagnostics []string) *core.AccessCheckResult {
	return &core.AccessCheckResult{
		Status:     core.AccessCheckStatusInconclusive,
		ReasonCode: reason,
		Items: []core.AccessCheckItem{
			{Id: core.AccessCheckItemConnect, Status: connectStatus},
			{Id: core.AccessCheckItemAuth, Status: core.AccessCheckItemStatusSkipped},
		},
		Diagnostics: diagnostics,
	}
}

func ConnectFailed(reason string, diagnostics []string) *core.AccessCheckResult {
	return &core.AccessCheckResult{
		Status:     core.AccessCheckStatusFailed,
		ReasonCode: reason,
		Items: []core.AccessCheckItem{
			{Id: core.AccessCheckItemConnect, Status: core.AccessCheckItemStatusFailed},
			{Id: core.AccessCheckItemAuth, Status: core.AccessCheckItemStatusSkipped},
		},
		Diagnostics: diagnostics,
	}
}

func CredentialVerdict(status int, code string, diagnostics []string) *core.AccessCheckResult {
	diag := append([]string{}, diagnostics...)
	if status > 0 {
		diag = append(diag, fmt.Sprintf("HTTP %d", status))
	}
	if code != "" {
		diag = append(diag, "code "+code)
	}
	switch {
	case authCode(code) || status == http.StatusUnauthorized:
		return AuthRejected(diag)
	case status == http.StatusForbidden || forbiddenCode(code):
		return Inconclusive(core.AccessCheckReasonProbeForbidden, core.AccessCheckItemStatusPassed, diag)
	case status >= 200 && status < 300 && code == "":
		return Passed(diag)
	default:
		if code == "" && status == 0 {
			diag = append(diag, "no credential verdict")
		}
		return Inconclusive(core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, diag)
	}
}

func FromTransport(err error, diagnostics []string) (*core.AccessCheckResult, bool) {
	if err == nil {
		return nil, false
	}
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return Inconclusive(core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, diagnostics), true
	case errors.Is(err, context.Canceled):
		return Inconclusive(core.AccessCheckReasonCanceled, core.AccessCheckItemStatusSkipped, diagnostics), true
	case isTLS(err):
		return ConnectFailed(core.AccessCheckReasonTlsFailed, append(append([]string{}, diagnostics...), "tls handshake failed")), true
	case isTimeout(err):
		return Inconclusive(core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, diagnostics), true
	case isConnect(err):
		return ConnectFailed(core.AccessCheckReasonConnectFailed, append(append([]string{}, diagnostics...), "tcp or dns failure")), true
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "context deadline exceeded") || strings.Contains(msg, "Client.Timeout exceeded") || strings.Contains(msg, "timeout awaiting response headers"):
		return Inconclusive(core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, diagnostics), true
	case strings.Contains(msg, "context canceled"):
		return Inconclusive(core.AccessCheckReasonCanceled, core.AccessCheckItemStatusSkipped, diagnostics), true
	case strings.Contains(msg, "tls: ") || strings.Contains(msg, "x509:"):
		return ConnectFailed(core.AccessCheckReasonTlsFailed, append(append([]string{}, diagnostics...), "tls handshake failed")), true
	case strings.Contains(msg, "dial tcp") || strings.Contains(msg, "connection refused") || strings.Contains(msg, "no such host") || strings.Contains(msg, "network is unreachable") || strings.Contains(msg, "connection reset by peer"):
		return ConnectFailed(core.AccessCheckReasonConnectFailed, append(append([]string{}, diagnostics...), "tcp or dns failure")), true
	default:
		return nil, false
	}
}

func Call(ctx context.Context, fn func() error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- fn()
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func Finish(err error, status int, code string, diagnostics []string, secrets ...string) *core.AccessCheckResult {
	if err != nil {
		if result, ok := FromTransport(err, diagnostics); ok {
			return Redact(result, secrets...)
		}
	}
	if err == nil && status == 0 {
		status = http.StatusOK
	}
	return Redact(CredentialVerdict(status, code, diagnostics), secrets...)
}

func Redact(result *core.AccessCheckResult, secrets ...string) *core.AccessCheckResult {
	if result == nil {
		return nil
	}
	out := *result
	out.Items = append([]core.AccessCheckItem(nil), result.Items...)
	out.Diagnostics = make([]string, len(result.Diagnostics))
	for i, line := range result.Diagnostics {
		out.Diagnostics[i] = scrub(line, secrets)
	}
	return &out
}

func scrub(text string, secrets []string) string {
	for _, secret := range secrets {
		if len(secret) < 4 {
			continue
		}
		text = strings.ReplaceAll(text, secret, "[redacted]")
	}
	return text
}

func authCode(code string) bool {
	return hasCodePrefix(code,
		"invalidaccesskeyid",
		"invalidaccesskey",
		"signaturedoesnotmatch",
		"incompletesignature",
		"authfailure",
		"invalidclienttokenid",
		"unrecognizedclientexception",
		"invalidsignatureexception",
		"signaturefailure",
		"expiredtoken",
		"invalidtoken",
		"tokenfailure",
		"invalidsecuritytoken",
		"invalidcredential",
		"invalidauthorization",
		"invalidhttpauthheader",
		"apigw.0301",
		"apigw.0307",
		"invalid_client",
		"invalid_grant",
		"unauthorized_client",
	)
}

func forbiddenCode(code string) bool {
	return hasCodePrefix(code,
		"accessdenied",
		"unauthorizedoperation",
		"forbidden",
		"optinrequired",
	)
}

func hasCodePrefix(code string, prefixes ...string) bool {
	c := strings.ToLower(strings.TrimSpace(code))
	if c == "" {
		return false
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(c, prefix) {
			return true
		}
	}
	return false
}

func isTLS(err error) bool {
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

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func isConnect(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	var dnsErr *net.DNSError
	return errors.As(err, &dnsErr)
}
