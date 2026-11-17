package actalisssl

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
)

const (
	eabKid     = "probe-kid-actalisssl"
	eabHmacKey = "probe-hmac-actalisssl"
)

func TestNew_Identity(t *testing.T) {
	checker := New()
	if checker.ProviderType() != string(domain.AccessProviderTypeActalisSSL) {
		t.Fatalf("provider type = %q, want %s", checker.ProviderType(), domain.AccessProviderTypeActalisSSL)
	}
	scope := checker.Scope()
	if len(scope) != 2 || scope[0] != core.AccessCheckItemConnect || scope[1] != core.AccessCheckItemAuth {
		t.Fatalf("scope = %v, want connect and auth", scope)
	}
}

func TestChecker_ProbesProductionDirectory(t *testing.T) {
	if got := New().(*checker).directoryEndpoint(); got != "https://acme-api.actalis.com/acme/directory" {
		t.Fatalf("directory URL = %q, want the certacme caDirUrls value", got)
	}
}

func TestChecker_Check_ConfigInvalid(t *testing.T) {
	cases := []struct {
		name   string
		config map[string]any
	}{
		{"nil config", nil},
		{"empty config", map[string]any{}},
		{"missing eabKid", map[string]any{"eabHmacKey": eabHmacKey}},
		{"blank eabKid", map[string]any{"eabKid": "   ", "eabHmacKey": eabHmacKey}},
		{"missing eabHmacKey", map[string]any{"eabKid": eabKid}},
		{"blank eabHmacKey", map[string]any{"eabKid": eabKid, "eabHmacKey": "   "}},
		{"malformed eabKid", map[string]any{"eabKid": 1, "eabHmacKey": eabHmacKey}},
		{"malformed eabHmacKey", map[string]any{"eabKid": eabKid, "eabHmacKey": true}},
	}

	checker := &checker{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := checker.Check(context.Background(), c.config)
			if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
				t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
			}
			if err != nil && (strings.Contains(err.Error(), eabKid) || strings.Contains(err.Error(), eabHmacKey)) {
				t.Fatal("config error leaked secret material")
			}
		})
	}
}

func TestChecker_Check_DirectoryReachableIsInconclusive(t *testing.T) {
	var gotMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(directoryBody()))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{baseURL: server.URL}).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, hostOf(server.URL), "were not verified")
	assertNoSecrets(t, result)
	if gotMethod != http.MethodGet {
		t.Fatalf("method = %q, want GET", gotMethod)
	}
}

func TestChecker_Check_InvalidDirectoryBody(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"json missing directory keys", `{"status":"ok"}`},
		{"html body", `<html>sign in</html>`},
		{"empty body", ``},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			server := httptest.NewServer(jsonStatus(http.StatusOK, c.body))
			t.Cleanup(server.Close)

			result, err := (&checker{baseURL: server.URL}).Check(context.Background(), validConfig())
			if err != nil {
				t.Fatal(err)
			}

			assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
			assertDiagnostics(t, result, hostOf(server.URL), "not a valid ACME directory")
			assertNoSecrets(t, result)
		})
	}
}

func TestChecker_Check_UnauthorizedIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusUnauthorized, `{"detail":"unauthorized"}`))
	t.Cleanup(server.Close)

	result, err := (&checker{baseURL: server.URL}).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, "HTTP 401")
	assertNoSecrets(t, result)
}

func TestChecker_Check_ForbiddenIsProbeForbidden(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusForbidden, `{"detail":"forbidden"}`))
	t.Cleanup(server.Close)

	result, err := (&checker{baseURL: server.URL}).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonProbeForbidden, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "HTTP 403")
	assertNoSecrets(t, result)
}

func TestChecker_Check_RefusedConnection(t *testing.T) {
	result, err := (&checker{baseURL: "http://" + closedAddr(t)}).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonConnectFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "tcp or dns failure")
	assertNoSecrets(t, result)
}

func TestChecker_Check_ConfigCannotOverrideEndpoint(t *testing.T) {
	fixed := httptest.NewServer(jsonStatus(http.StatusOK, directoryBody()))
	t.Cleanup(fixed.Close)

	decoyHits := 0
	decoy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoyHits++
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(decoy.Close)

	config := validConfig()
	config["endpoint"] = decoy.URL
	config["baseUrl"] = decoy.URL
	config["directoryUrl"] = decoy.URL

	result, err := (&checker{baseURL: fixed.URL}).Check(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	if decoyHits != 0 {
		t.Fatalf("config-supplied endpoint was contacted %d times", decoyHits)
	}
}

func validConfig() map[string]any {
	return map[string]any{
		"eabKid":     eabKid,
		"eabHmacKey": eabHmacKey,
	}
}

func directoryBody() string {
	return `{"newAccount":"https://acme-api.actalis.com/acme/new-account","newNonce":"https://acme-api.actalis.com/acme/new-nonce","newOrder":"https://acme-api.actalis.com/acme/new-order"}`
}

func jsonStatus(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func hostOf(raw string) string {
	return strings.TrimPrefix(strings.TrimPrefix(raw, "https://"), "http://")
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
	for _, secret := range []string{eabKid, eabHmacKey} {
		if strings.Contains(joined, secret) {
			t.Fatal("diagnostics leaked secret material")
		}
	}
}
