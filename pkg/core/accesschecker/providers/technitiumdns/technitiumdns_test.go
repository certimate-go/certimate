package technitiumdns

import (
	"context"
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

const valApiToken = "probe-technitium-token"

func TestNew_Identity(t *testing.T) {
	checker := New()
	if checker.ProviderType() != string(domain.AccessProviderTypeTechnitiumDNS) {
		t.Fatalf("provider type = %q, want %s", checker.ProviderType(), domain.AccessProviderTypeTechnitiumDNS)
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
		{"missing serverUrl", map[string]any{"apiToken": valApiToken}},
		{"blank serverUrl", map[string]any{"serverUrl": "   ", "apiToken": valApiToken}},
		{"invalid serverUrl", map[string]any{"serverUrl": "technitium.local", "apiToken": valApiToken}},
		{"wrong scheme", map[string]any{"serverUrl": "ftp://technitium.local", "apiToken": valApiToken}},
		{"missing apiToken", map[string]any{"serverUrl": "http://technitium.local"}},
		{"blank apiToken", map[string]any{"serverUrl": "http://technitium.local", "apiToken": "   "}},
		{"malformed apiToken", map[string]any{"serverUrl": "http://technitium.local", "apiToken": 1}},
	}

	checker := &checker{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := checker.Check(context.Background(), c.config)
			if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
				t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
			}
			if err != nil && strings.Contains(err.Error(), valApiToken) {
				t.Fatal("config error leaked secret material")
			}
		})
	}
}

func TestChecker_Check_Passes(t *testing.T) {
	var gotMethod, gotPath, gotForm string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotForm = string(body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","response":{"zones":[],"totalZones":0}}`))
	}))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), configFor(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertDiagnostics(t, result, hostOf(server.URL), "HTTP 200", "credential accepted")
	assertNoSecrets(t, result)
	if gotMethod != http.MethodPost {
		t.Fatalf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/api/zones/list" {
		t.Fatalf("path = %q, want /api/zones/list", gotPath)
	}
	if !strings.Contains(gotForm, "token="+valApiToken) {
		t.Fatalf("probe did not send the api token: form=%s", gotForm)
	}
}

func TestChecker_Check_InsecureConnectionsSkipTlsVerify(t *testing.T) {
	server := httptest.NewTLSServer(jsonStatus(http.StatusOK, `{"status":"ok","response":{"zones":[],"totalZones":0}}`))
	t.Cleanup(server.Close)

	config := validConfig()
	config["allowInsecureConnections"] = true
	config["serverUrl"] = server.URL

	result, err := (&checker{}).Check(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertNoSecrets(t, result)
}

func TestChecker_Check_InvalidTokenIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusOK, `{"status":"invalid-token","errorMessage":"token is not valid"}`))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), configFor(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, "invalid-token")
	assertNoSecrets(t, result)
}

func TestChecker_Check_UnauthorizedStatusIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusUnauthorized, `{"status":"invalid-token","errorMessage":"token is not valid"}`))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), configFor(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertNoSecrets(t, result)
}

func TestChecker_Check_ForbiddenIsInconclusive(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusForbidden, `{"status":"error","errorMessage":"forbidden"}`))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), configFor(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonProbeForbidden, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_ApiErrorIsRemoteError(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusOK, `{"status":"error","errorMessage":"zone not found"}`))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), configFor(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "zone not found")
	assertNoSecrets(t, result)
}

func TestChecker_Check_RefusedConnection(t *testing.T) {
	result, err := (&checker{}).Check(context.Background(), configFor("http://"+closedAddr(t)))
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

	result, err := checkerFor(t, server.URL).Check(context.Background(), configFor(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonTlsFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "tls handshake failed")
	assertNoSecrets(t, result)
}

func TestChecker_Check_TimeoutIsInconclusive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	}))
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result, err := checkerFor(t, server.URL).Check(ctx, configFor(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func validConfig() map[string]any {
	return map[string]any{
		"serverUrl": "http://technitium.local",
		"apiToken":  valApiToken,
	}
}

func checkerFor(t *testing.T, raw string) *checker {
	t.Helper()
	return &checker{}
}

func configFor(raw string) map[string]any {
	config := validConfig()
	config["serverUrl"] = raw
	return config
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

func jsonStatus(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func hostOf(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return u.Host
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
	if strings.Contains(joined, valApiToken) {
		t.Fatal("diagnostics leaked secret material")
	}
}
