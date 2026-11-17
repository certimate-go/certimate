package cpanel

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/certimate-go/certimate/pkg/core"
)

const apiToken = "cpanel-api-token-secret"

func TestChecker_Check_ConfigInvalid(t *testing.T) {
	cases := []struct {
		name   string
		config map[string]any
	}{
		{"empty config", map[string]any{}},
		{"missing serverUrl", map[string]any{"username": "root", "apiToken": apiToken}},
		{"blank serverUrl", map[string]any{"serverUrl": "  ", "username": "root", "apiToken": apiToken}},
		{"serverUrl without scheme", map[string]any{"serverUrl": "127.0.0.1:2083", "username": "root", "apiToken": apiToken}},
		{"missing username", map[string]any{"serverUrl": "https://127.0.0.1:2083", "apiToken": apiToken}},
		{"missing apiToken", map[string]any{"serverUrl": "https://127.0.0.1:2083", "username": "root"}},
		{"blank apiToken", map[string]any{"serverUrl": "https://127.0.0.1:2083", "username": "root", "apiToken": "  "}},
		{"malformed allowInsecureConnections", map[string]any{"serverUrl": "https://127.0.0.1:2083", "username": "root", "apiToken": apiToken, "allowInsecureConnections": "yes"}},
	}

	checker := &checker{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := checker.Check(context.Background(), c.config)
			if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
				t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
			}
			if strings.Contains(err.Error(), apiToken) {
				t.Fatalf("config error leaked token material: %v", err)
			}
		})
	}
}

func TestChecker_Check_Passes(t *testing.T) {
	var got probeRecord
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = probeRecord{method: r.Method, path: r.URL.Path, auth: r.Header.Get("Authorization")}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":1,"data":[]}`))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{}).Check(context.Background(), map[string]any{
		"serverUrl": server.URL,
		"username":  "root",
		"apiToken":  apiToken,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertDiagnostics(t, result, hostOf(server.URL), "HTTP 200", "credential accepted")
	assertNoSecrets(t, result)
	if got.method != http.MethodGet {
		t.Fatalf("method = %q, want GET", got.method)
	}
	if got.path != "/execute/VersionControl/get_repositories" {
		t.Fatalf("path = %q, want /execute/VersionControl/get_repositories", got.path)
	}
	if got.auth != "cpanel root:"+apiToken {
		t.Fatalf("authorization header did not use the cpanel username:token form")
	}
}

func TestChecker_Check_UnauthorizedIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{}).Check(context.Background(), map[string]any{
		"serverUrl": server.URL,
		"username":  "root",
		"apiToken":  apiToken,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, "HTTP 401")
	assertNoSecrets(t, result)
}

func TestChecker_Check_ForbiddenIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{}).Check(context.Background(), map[string]any{
		"serverUrl": server.URL,
		"username":  "root",
		"apiToken":  apiToken,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertNoSecrets(t, result)
}

func TestChecker_Check_ServerErrorIsRemoteError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream panicked "+apiToken, http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{}).Check(context.Background(), map[string]any{
		"serverUrl": server.URL,
		"username":  "root",
		"apiToken":  apiToken,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_RefusedConnection(t *testing.T) {
	result, err := (&checker{}).Check(context.Background(), map[string]any{
		"serverUrl": "http://" + closedAddr(t),
		"username":  "root",
		"apiToken":  apiToken,
	})
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

	result, err := (&checker{}).Check(context.Background(), map[string]any{
		"serverUrl": server.URL,
		"username":  "root",
		"apiToken":  apiToken,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonTlsFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "tls handshake failed")
	assertNoSecrets(t, result)
}

func TestChecker_Check_InsecureTLSPasses(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{}).Check(context.Background(), map[string]any{
		"serverUrl":                server.URL,
		"username":                 "root",
		"apiToken":                 apiToken,
		"allowInsecureConnections": true,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertNoSecrets(t, result)
}

func TestChecker_Check_TimeoutIsInconclusive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result, err := (&checker{}).Check(ctx, map[string]any{
		"serverUrl": server.URL,
		"username":  "root",
		"apiToken":  apiToken,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

type probeRecord struct {
	method string
	path   string
	auth   string
}

func hostOf(raw string) string {
	return strings.TrimPrefix(strings.TrimPrefix(raw, "https://"), "http://")
}

func closedAddr(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()
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
	if strings.Contains(joined, apiToken) {
		t.Errorf("diagnostics leaked token material: %v", got.Diagnostics)
	}
}
