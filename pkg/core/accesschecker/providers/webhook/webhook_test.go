package webhook

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/certimate-go/certimate/pkg/core"
)

const (
	webhookToken = "webhook-secret-token"
	webhookData  = `{"cert":"${CERTIMATE_DEPLOYER_CERTIFICATE}"}`
)

func TestChecker_Check_ConfigInvalid(t *testing.T) {
	cases := []struct {
		name   string
		config map[string]any
	}{
		{"empty config", map[string]any{}},
		{"missing url", map[string]any{"method": "POST"}},
		{"blank url", map[string]any{"url": "  "}},
		{"url without scheme", map[string]any{"url": "127.0.0.1:9"}},
		{"unsupported scheme", map[string]any{"url": "ftp://127.0.0.1/hook"}},
		{"invalid method", map[string]any{"url": "http://127.0.0.1:9/hook", "method": "HEAD"}},
		{"malformed method type", map[string]any{"url": "http://127.0.0.1:9/hook", "method": 1}},
		{"malformed headers", map[string]any{"url": "http://127.0.0.1:9/hook", "headers": "not-a-header " + webhookToken}},
		{"malformed json headers", map[string]any{"url": "http://127.0.0.1:9/hook", "headers": `{"Authorization":` + webhookToken}},
	}

	checker := &checker{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := checker.Check(context.Background(), c.config)
			if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
				t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
			}
			if strings.Contains(err.Error(), webhookToken) {
				t.Fatalf("config error leaked token material: %v", err)
			}
		})
	}
}

func TestChecker_Check_PassesAndSendsConfiguredRequest(t *testing.T) {
	var got methodRecord
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		got = methodRecord{
			method: r.Method,
			auth:   r.Header.Get("Authorization"),
			probe:  r.Header.Get("X-Probe"),
			body:   string(body),
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{}).Check(context.Background(), map[string]any{
		"url":     server.URL + "/hook",
		"method":  "put",
		"headers": "Authorization: Bearer " + webhookToken + "\r\nX-Probe: 1",
		"data":    webhookData,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertDiagnostics(t, result, hostOf(server.URL), "PUT", "HTTP 204", "credential accepted")
	assertNoSecrets(t, result)
	if got.method != http.MethodPut {
		t.Fatalf("method = %q, want PUT", got.method)
	}
	if got.auth != "Bearer "+webhookToken {
		t.Fatalf("authorization header was not forwarded")
	}
	if got.probe != "1" {
		t.Fatalf("X-Probe = %q, want 1", got.probe)
	}
	if got.body != webhookData {
		t.Fatalf("body = %q, want the configured data with placeholders left intact", got.body)
	}
}

func TestChecker_Check_JSONHeadersAndDefaultMethod(t *testing.T) {
	var got methodRecord
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = methodRecord{method: r.Method, auth: r.Header.Get("Authorization")}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{}).Check(context.Background(), map[string]any{
		"url":     server.URL,
		"headers": `{"Authorization":"Bearer ` + webhookToken + `"}`,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertNoSecrets(t, result)
	if got.method != http.MethodPost {
		t.Fatalf("method = %q, want POST", got.method)
	}
	if got.auth != "Bearer "+webhookToken {
		t.Fatalf("json authorization header was not forwarded")
	}
}

func TestChecker_Check_UnauthorizedIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{}).Check(context.Background(), map[string]any{"url": server.URL})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, "HTTP 401")
}

func TestChecker_Check_ForbiddenIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{}).Check(context.Background(), map[string]any{"url": server.URL})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
}

func TestChecker_Check_ServerErrorIsRemoteError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream panicked "+webhookToken, http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{}).Check(context.Background(), map[string]any{"url": server.URL})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_RedirectIsRemoteError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "http://127.0.0.1:1/elsewhere")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{}).Check(context.Background(), map[string]any{"url": server.URL})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "HTTP 302")
}

func TestChecker_Check_RefusedConnection(t *testing.T) {
	result, err := (&checker{}).Check(context.Background(), map[string]any{
		"url": "http://" + closedAddr(t) + "/hook",
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonConnectFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "tcp or dns failure")
}

func TestChecker_Check_TlsFailure(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("the tls fixture must fail the handshake before any handler runs")
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{}).Check(context.Background(), map[string]any{"url": server.URL})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonTlsFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "tls handshake failed")
}

func TestChecker_Check_InsecureTLSPasses(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{}).Check(context.Background(), map[string]any{
		"url":                      server.URL,
		"allowInsecureConnections": true,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
}

func TestChecker_Check_TimeoutIsInconclusive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result, err := (&checker{}).Check(ctx, map[string]any{"url": server.URL})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, core.AccessCheckItemStatusSkipped)
}

type methodRecord struct {
	method string
	auth   string
	probe  string
	body   string
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
	for _, secret := range []string{webhookToken, webhookData, "${CERTIMATE_DEPLOYER_CERTIFICATE}"} {
		if strings.Contains(joined, secret) {
			t.Errorf("diagnostics leaked secret or payload material: %v", got.Diagnostics)
		}
	}
}
