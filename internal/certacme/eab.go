package certacme

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/certimate-go/certimate/internal/domain"
	xhttp "github.com/certimate-go/certimate/pkg/utils/http"
)

const (
	zerosslEABProviderId       = "zerossl"
	eabAcquireMaxResponseBytes = 64 * 1024
	eabAcquireRequestTimeout   = 30 * time.Second
)

var zerosslEABEndpoint = "https://api.zerossl.com/acme/eab-credentials-email"

const (
	eabAcquireStageRequest     = "request"
	eabAcquireStageTransport   = "transport"
	eabAcquireStageTimeout     = "timeout"
	eabAcquireStageCanceled    = "canceled"
	eabAcquireStageRedirect    = "redirect"
	eabAcquireStageHTTPStatus  = "http_status"
	eabAcquireStageResponse    = "response"
	eabAcquireStageBusiness    = "business"
	eabAcquireStageCredentials = "credentials"
)

var (
	errEABRedirectRefused  = errors.New("the acme eab acquisition does not follow redirects")
	eabUpstreamCodePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:]{0,31}$`)
)

type eabCredential struct {
	Kid            string
	HmacKeyEncoded string
}

type ACMEEABProvider interface {
	AcquireEAB(ctx context.Context, email string) (*eabCredential, error)
	Reusable() bool
}

var acmeEABProviders = map[domain.CAProviderType]ACMEEABProvider{
	domain.CAProviderTypeZeroSSL: zerosslACMEEABProvider{},
}

func findACMEEABProvider(ca domain.CAProviderType) ACMEEABProvider {
	return acmeEABProviders[ca]
}

type eabAcquireError struct {
	Stage        string
	ProviderId   string
	HTTPStatus   int
	UpstreamCode string

	cause error
}

func (e *eabAcquireError) Error() string {
	var sb strings.Builder
	sb.WriteString("failed to acquire the acme eab credentials (stage=")
	sb.WriteString(e.Stage)
	sb.WriteString(", provider=")
	sb.WriteString(e.ProviderId)
	if e.HTTPStatus > 0 {
		sb.WriteString(", http_status=")
		sb.WriteString(strconv.Itoa(e.HTTPStatus))
	}
	if e.UpstreamCode != "" {
		sb.WriteString(", code=")
		sb.WriteString(e.UpstreamCode)
	}
	sb.WriteString(")")

	return sb.String()
}

func (e *eabAcquireError) Unwrap() error {
	return e.cause
}

func newEABHTTPClient() *http.Client {
	return &http.Client{
		Transport: xhttp.NewDefaultTransport(),
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return errEABRedirectRefused
		},
	}
}

func classifyEABTransportError(providerId string, err error) *eabAcquireError {
	acquireErr := &eabAcquireError{ProviderId: providerId, cause: err}
	switch {
	case errors.Is(err, errEABRedirectRefused):
		acquireErr.Stage = eabAcquireStageRedirect
	case errors.Is(err, context.Canceled):
		acquireErr.Stage = eabAcquireStageCanceled
	case errors.Is(err, context.DeadlineExceeded):
		acquireErr.Stage = eabAcquireStageTimeout
	default:
		acquireErr.Stage = eabAcquireStageTransport
	}

	return acquireErr
}

func readEABResponseBody(body io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(body, eabAcquireMaxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read the response body: %w", err)
	}
	if len(data) > eabAcquireMaxResponseBytes {
		return nil, fmt.Errorf("the response body exceeds %d bytes", eabAcquireMaxResponseBytes)
	}

	return data, nil
}

func sanitizeEABUpstreamErrorCode(value any) string {
	switch v := value.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		code := strconv.FormatFloat(v, 'f', -1, 64)
		if len(code) > 64 {
			return ""
		}
		return code
	case string:
		if eabUpstreamCodePattern.MatchString(v) {
			return v
		}
	}

	return ""
}

func normalizeEABHmacKeyEncoded(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(strings.ReplaceAll(value, "+", "-"), "/", "_")
	return strings.TrimRight(value, "=")
}

type zerosslEABResponse struct {
	Success    any              `json:"success"`
	EABKid     string           `json:"eab_kid"`
	EABHmacKey string           `json:"eab_hmac_key"`
	Error      *zerosslEABError `json:"error"`
}

type zerosslEABError struct {
	Code any `json:"code"`
}

type zerosslACMEEABProvider struct{}

func (zerosslACMEEABProvider) Reusable() bool {
	return true
}

func (zerosslACMEEABProvider) AcquireEAB(ctx context.Context, email string) (*eabCredential, error) {
	if strings.TrimSpace(email) == "" {
		return nil, &eabAcquireError{Stage: eabAcquireStageRequest, ProviderId: zerosslEABProviderId, cause: fmt.Errorf("the registration email is empty")}
	}

	requestCtx, cancel := context.WithTimeout(ctx, eabAcquireRequestTimeout)
	defer cancel()

	form := url.Values{"email": []string{email}}
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, zerosslEABEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, &eabAcquireError{Stage: eabAcquireStageRequest, ProviderId: zerosslEABProviderId, cause: err}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := newEABHTTPClient().Do(req)
	if err != nil {
		return nil, classifyEABTransportError(zerosslEABProviderId, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		upstreamCode := ""
		if data, readErr := readEABResponseBody(resp.Body); readErr == nil {
			var document zerosslEABResponse
			if json.Unmarshal(data, &document) == nil && document.Error != nil {
				upstreamCode = sanitizeEABUpstreamErrorCode(document.Error.Code)
			}
		}

		return nil, &eabAcquireError{
			Stage:        eabAcquireStageHTTPStatus,
			ProviderId:   zerosslEABProviderId,
			HTTPStatus:   resp.StatusCode,
			UpstreamCode: upstreamCode,
		}
	}

	data, err := readEABResponseBody(resp.Body)
	if err != nil {
		return nil, &eabAcquireError{Stage: eabAcquireStageResponse, ProviderId: zerosslEABProviderId, HTTPStatus: resp.StatusCode, cause: err}
	}

	var document zerosslEABResponse
	if err := json.Unmarshal(data, &document); err != nil {
		return nil, &eabAcquireError{Stage: eabAcquireStageResponse, ProviderId: zerosslEABProviderId, HTTPStatus: resp.StatusCode, cause: fmt.Errorf("the response body is not valid json")}
	}

	switch success := document.Success.(type) {
	case bool:
		if success != true {
			return nil, &eabAcquireError{Stage: eabAcquireStageBusiness, ProviderId: zerosslEABProviderId, HTTPStatus: resp.StatusCode, UpstreamCode: zerosslEABErrorCode(document)}
		}
	case float64:
		if success != 1 {
			return nil, &eabAcquireError{Stage: eabAcquireStageBusiness, ProviderId: zerosslEABProviderId, HTTPStatus: resp.StatusCode, UpstreamCode: zerosslEABErrorCode(document)}
		}
	default:
		return nil, &eabAcquireError{Stage: eabAcquireStageBusiness, ProviderId: zerosslEABProviderId, HTTPStatus: resp.StatusCode, UpstreamCode: zerosslEABErrorCode(document)}
	}

	kid := strings.TrimSpace(document.EABKid)
	hmacKeyEncoded := normalizeEABHmacKeyEncoded(document.EABHmacKey)
	hmacKeyDecoded, hmacKeyDecodeErr := base64.RawURLEncoding.DecodeString(hmacKeyEncoded)
	if kid == "" || hmacKeyEncoded == "" || hmacKeyDecodeErr != nil || len(hmacKeyDecoded) == 0 {
		return nil, &eabAcquireError{Stage: eabAcquireStageCredentials, ProviderId: zerosslEABProviderId, HTTPStatus: resp.StatusCode}
	}

	return &eabCredential{Kid: kid, HmacKeyEncoded: hmacKeyEncoded}, nil
}

func zerosslEABErrorCode(document zerosslEABResponse) string {
	if document.Error == nil {
		return ""
	}

	return sanitizeEABUpstreamErrorCode(document.Error.Code)
}
