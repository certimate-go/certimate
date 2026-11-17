package akamai

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
)

const (
	valHost         = "akab-probe.edgegrid.akamaiapis.net"
	valClientToken  = "probe-akamai-client-token"
	valClientSecret = "probe-akamai-client-secret"
	valAccessToken  = "probe-akamai-access-token"
)

func TestNew_Identity(t *testing.T) {
	checker := New()
	if checker.ProviderType() != string(domain.AccessProviderTypeAkamai) {
		t.Fatalf("provider type = %q, want %s", checker.ProviderType(), domain.AccessProviderTypeAkamai)
	}
	scope := checker.Scope()
	if len(scope) != 2 || scope[0] != core.AccessCheckItemConnect || scope[1] != core.AccessCheckItemAuth {
		t.Fatalf("scope = %v, want connect and auth", scope)
	}
}

func TestChecker_Check_ConfigInvalid(t *testing.T) {
	cases := []struct {
		name   string
		config map[string]any
	}{
		{"nil config", nil},
		{"empty config", map[string]any{}},
		{"missing host", map[string]any{"clientToken": valClientToken, "clientSecret": valClientSecret, "accessToken": valAccessToken}},
		{"host with scheme", map[string]any{"host": "https://" + valHost, "clientToken": valClientToken, "clientSecret": valClientSecret, "accessToken": valAccessToken}},
		{"host with path", map[string]any{"host": valHost + "/path", "clientToken": valClientToken, "clientSecret": valClientSecret, "accessToken": valAccessToken}},
		{"missing clientToken", map[string]any{"host": valHost, "clientSecret": valClientSecret, "accessToken": valAccessToken}},
		{"missing clientSecret", map[string]any{"host": valHost, "clientToken": valClientToken, "accessToken": valAccessToken}},
		{"missing accessToken", map[string]any{"host": valHost, "clientToken": valClientToken, "clientSecret": valClientSecret}},
		{"malformed host", map[string]any{"host": 1, "clientToken": valClientToken, "clientSecret": valClientSecret, "accessToken": valAccessToken}},
	}

	checker := &checker{client: insecureClient()}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := checker.Check(context.Background(), c.config)
			if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
				t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
			}
			if err != nil && (strings.Contains(err.Error(), valClientSecret) || strings.Contains(err.Error(), valAccessToken) || strings.Contains(err.Error(), valClientToken)) {
				t.Fatal("config error leaked secret material")
			}
		})
	}
}

func TestChecker_Check_ListZonesSignedRequestPasses(t *testing.T) {
	var gotPath, gotQuery, gotAuth string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"zones":[{"zone":"example.com","type":"PRIMARY"}],"page":1,"pageSize":1}`))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{client: insecureClient()}).Check(context.Background(), validConfig(hostOf(server.URL)))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertDiagnostics(t, result, hostOf(server.URL), "HTTP 200", "credential accepted")
	assertNoSecrets(t, result)
	if gotPath != "/config-dns/v2/zones" {
		t.Fatalf("path = %q, want /config-dns/v2/zones", gotPath)
	}
	if !strings.Contains(gotQuery, "page=1") || !strings.Contains(gotQuery, "pageSize=1") {
		t.Fatalf("query = %q, want a single-page zone list", gotQuery)
	}
	if !strings.HasPrefix(gotAuth, "EG1-HMAC-SHA256 ") {
		t.Fatalf("authorization = %q, want an EdgeGrid signed header", gotAuth)
	}
}

func TestChecker_Check_UnauthorizedIsAuthRejected(t *testing.T) {
	server := httptest.NewTLSServer(jsonStatus(http.StatusUnauthorized, `{"type":"internal-server-error","title":"Unauthorized","detail":"Invalid credentials"}`))
	t.Cleanup(server.Close)

	result, err := (&checker{client: insecureClient()}).Check(context.Background(), validConfig(hostOf(server.URL)))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, "HTTP 401")
	assertNoSecrets(t, result)
}

func TestChecker_Check_ForbiddenIsInconclusive(t *testing.T) {
	server := httptest.NewTLSServer(jsonStatus(http.StatusForbidden, `{"title":"Forbidden","detail":"no access to this API"}`))
	t.Cleanup(server.Close)

	result, err := (&checker{client: insecureClient()}).Check(context.Background(), validConfig(hostOf(server.URL)))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonProbeForbidden, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_ServerErrorIsRemoteError(t *testing.T) {
	server := httptest.NewTLSServer(jsonStatus(http.StatusInternalServerError, `{"title":"Internal Server Error"}`))
	t.Cleanup(server.Close)

	result, err := (&checker{client: insecureClient()}).Check(context.Background(), validConfig(hostOf(server.URL)))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_RefusedConnection(t *testing.T) {
	result, err := (&checker{client: insecureClient()}).Check(context.Background(), validConfig(closedAddr(t)))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonConnectFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "tcp or dns failure")
	assertNoSecrets(t, result)
}

func TestChecker_Check_TlsFailure(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("the tls fixture must fail the handshake before any handler runs")
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{}).Check(context.Background(), validConfig(hostOf(server.URL)))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonTlsFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "tls handshake failed")
	assertNoSecrets(t, result)
}

func TestChecker_Check_TimeoutIsInconclusive(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	}))
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result, err := (&checker{client: insecureClient()}).Check(ctx, validConfig(hostOf(server.URL)))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func validConfig(host string) map[string]any {
	return map[string]any{
		"host":         host,
		"clientToken":  valClientToken,
		"clientSecret": valClientSecret,
		"accessToken":  valAccessToken,
	}
}

func insecureClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func closedAddr(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	listener.Close()
	return addr
}

func hostOf(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return u.Host
}

func jsonStatus(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func assertResult(t *testing.T, got *core.AccessCheckResult, wantStatus string, wantReason string, wantConnect string, wantAuth string) {
	t.Helper()
	if got == nil {
		t.Fatal("expected a result, got nil")
	}
	if got.Status != wantStatus {
		t.Errorf("status = %q, want %q", got.Status, wantStatus)
	}
	if got.ReasonCode != wantReason {
		t.Errorf("reasonCode = %q, want %q", got.ReasonCode, wantReason)
	}
	if len(got.Items) != 2 {
		t.Fatalf("items = %+v, want exactly connect and auth", got.Items)
	}
	if got.Items[0].Id != core.AccessCheckItemConnect || got.Items[0].Status != wantConnect {
		t.Errorf("connect item = %+v, want status %q", got.Items[0], wantConnect)
	}
	if got.Items[1].Id != core.AccessCheckItemAuth || got.Items[1].Status != wantAuth {
		t.Errorf("auth item = %+v, want status %q", got.Items[1], wantAuth)
	}
}

func assertDiagnostics(t *testing.T, got *core.AccessCheckResult, mustContain ...string) {
	t.Helper()
	joined := strings.Join(got.Diagnostics, "\n")
	for _, substr := range mustContain {
		if !strings.Contains(joined, substr) {
			t.Errorf("diagnostics missing %q: %v", substr, got.Diagnostics)
		}
	}
}

func assertNoSecrets(t *testing.T, got *core.AccessCheckResult) {
	t.Helper()
	joined := strings.Join(got.Diagnostics, "\n")
	for _, secret := range []string{valClientToken, valClientSecret, valAccessToken} {
		if strings.Contains(joined, secret) {
			t.Fatal("diagnostics leaked secret material")
		}
	}
}
