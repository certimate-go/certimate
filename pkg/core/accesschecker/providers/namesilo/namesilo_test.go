package namesilo

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

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
)

const secret = "probe-secret-namesilo"

func TestNew_Identity(t *testing.T) {
	checker := New()
	if checker.ProviderType() != string(domain.AccessProviderTypeNameSilo) {
		t.Fatalf("provider type = %q, want %s", checker.ProviderType(), domain.AccessProviderTypeNameSilo)
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
		{"missing secret", map[string]any{"note": secret}},
		{"blank secret", map[string]any{"apiKey": "   ", "note": secret}},
		{"malformed secret", map[string]any{"apiKey": 1, "note": secret}},
	}

	checker := &checker{baseURL: "http://127.0.0.1:1"}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := checker.Check(context.Background(), c.config)
			if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
				t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
			}
			if err != nil && strings.Contains(err.Error(), secret) {
				t.Fatalf("config error leaked secret material: %v", err)
			}
		})
	}
}

func TestChecker_Check_Passes(t *testing.T) {
	var got probeRecord
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		got = probeRecord{
			method: r.Method,
			path:   r.URL.Path,
			query:  r.URL.Query().Get("key"),
			body:   string(body),
		}
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<namesilo><reply><code>300</code><detail>" + secret + "</detail></reply></namesilo>"))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{baseURL: server.URL}).Check(context.Background(), configWith(secret))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertDiagnostics(t, result, hostOf(server.URL), "HTTP 200", "reply code 300", "credential accepted")
	assertNoSecrets(t, result)
	if got.method != http.MethodGet {
		t.Fatalf("method = %q, want GET", got.method)
	}
	if got.path != "/listDomains" {
		t.Fatalf("path = %q, want /listDomains", got.path)
	}
	if got.query != secret {
		t.Fatal("query key did not match the configured secret")
	}
}

func TestChecker_Check_ReplyCodes(t *testing.T) {
	cases := []struct {
		code        string
		wantStatus  string
		wantReason  string
		wantConnect string
		wantAuth    string
	}{
		{"301", core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed},
		{"302", core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed},
		{"110", core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed},
		{"250", core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped},
	}

	for _, c := range cases {
		t.Run(c.code, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("<namesilo><reply><code>" + c.code + "</code><detail>" + secret + "</detail></reply></namesilo>"))
			}))
			t.Cleanup(server.Close)

			result, err := (&checker{baseURL: server.URL}).Check(context.Background(), configWith(secret))
			if err != nil {
				t.Fatal(err)
			}
			assertResult(t, result, c.wantStatus, c.wantReason, c.wantConnect, c.wantAuth)
			assertDiagnostics(t, result, "reply code "+c.code)
			assertNoSecrets(t, result)
		})
	}
}

func TestChecker_Check_UnauthorizedIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(secret))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{baseURL: server.URL}).Check(context.Background(), configWith(secret))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, "HTTP 401")
	assertNoSecrets(t, result)
}

func TestChecker_Check_ServerErrorIgnoresSuccessBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("<namesilo><reply><code>300</code><detail>" + secret + "</detail></reply></namesilo>"))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{baseURL: server.URL}).Check(context.Background(), configWith(secret))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_RefusedConnection(t *testing.T) {
	result, err := (&checker{baseURL: "http://" + closedAddr(t)}).Check(context.Background(), configWith(secret))
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

	result, err := (&checker{baseURL: server.URL}).Check(context.Background(), configWith(secret))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonTlsFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "tls handshake failed")
	assertNoSecrets(t, result)
}

func TestChecker_Check_TimeoutIsInconclusive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result, err := (&checker{baseURL: server.URL}).Check(ctx, configWith(secret))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_ConfigCannotOverrideEndpoint(t *testing.T) {
	var hits int
	fixed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<namesilo><reply><code>300</code><detail>success</detail></reply></namesilo>"))
	}))
	t.Cleanup(fixed.Close)

	decoyHits := 0
	decoy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoyHits++
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(decoy.Close)

	config := configWith(secret)
	config["baseUrl"] = decoy.URL
	config["endpoint"] = decoy.URL
	config["apiUrl"] = decoy.URL

	result, err := (&checker{baseURL: fixed.URL}).Check(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertNoSecrets(t, result)
	if decoyHits != 0 {
		t.Fatal("config-supplied endpoint keys must never be contacted")
	}
	if hits != 1 {
		t.Fatalf("fixed probe hits = %d, want 1", hits)
	}
}

func TestProbeEndpoint_ProductionURL(t *testing.T) {
	endpoint, err := probeEndpoint(apiBaseURL, secret)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := endpoint.Scheme+"://"+endpoint.Host+endpoint.Path, "https://www.namesilo.com/api/listDomains"; got != want {
		t.Fatalf("probe url = %q, want %q", got, want)
	}
	if endpoint.Query().Get("version") != "1" || endpoint.Query().Get("type") != "xml" {
		t.Fatal("production probe query is missing version or type")
	}
	if endpoint.Query().Get("key") != secret {
		t.Fatal("production probe query key did not match the configured secret")
	}
}

func configWith(value string) map[string]any {
	return map[string]any{"apiKey": value}
}

type probeRecord struct {
	method string
	path   string
	auth   string
	query  string
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
	if strings.Contains(joined, secret) {
		t.Errorf("diagnostics leaked secret material: %v", got.Diagnostics)
	}
}
