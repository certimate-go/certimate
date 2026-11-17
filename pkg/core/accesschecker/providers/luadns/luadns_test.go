package luadns

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
	"encoding/base64"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
)

const (
	valLeak = "probe-leak-luadns"
	valUsername = "probe-luadns-username"
	valApiKey = "probe-luadns-apiKey"
)

func TestNew_Identity(t *testing.T) {
	checker := New()
	if checker.ProviderType() != string(domain.AccessProviderTypeLuaDNS) {
		t.Fatalf("provider type = %q, want %s", checker.ProviderType(), domain.AccessProviderTypeLuaDNS)
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
		{"missing username", map[string]any{"apiKey": valApiKey}},
		{"blank username", map[string]any{"username": "   ", "apiKey": valApiKey}},
		{"malformed username", map[string]any{"username": 1, "apiKey": valApiKey}},
		{"missing apiKey", map[string]any{"username": valUsername}},
		{"blank apiKey", map[string]any{"apiKey": "   ", "username": valUsername}},
		{"malformed apiKey", map[string]any{"apiKey": 1, "username": valUsername}},
		{"missing secret", map[string]any{"note": valLeak}},
	}

	checker := &checker{baseURL: "http://127.0.0.1:1"}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := checker.Check(context.Background(), c.config)
			if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
				t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
			}
			if err != nil && containsSecret(err.Error()) {
				t.Fatalf("config error leaked secret material: %v", err)
			}
		})
	}
}

func TestChecker_Check_Passes(t *testing.T) {
	all := serve(t, http.StatusOK, passBody())
	result, err := newTestChecker(t, all.server).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertDiagnostics(t, result, hostOf(all.server.URL), "HTTP 200", "credential accepted", "permissions on specific resources were not checked")
	assertNoSecrets(t, result)
	if len(all.got) != 1 {
		t.Fatalf("probe hits = %d, want 1", len(all.got))
	}
	got := all.got[0]
	if got.method != http.MethodGet {
		t.Fatalf("method = %q", got.method)
	}
	if got.path != "/v1/zones" {
		t.Fatalf("path = %q, want /v1/zones", got.path)
	}
	user, passwd, ok := basicAuth(got.header.Get("Authorization"))
	if !ok || user != valUsername || passwd != valApiKey {
		t.Fatal("basic auth did not match the configured username and api key")
	}
}



func TestChecker_Check_UnauthorizedIsAuthRejected(t *testing.T) {
	all := serve(t, http.StatusUnauthorized, valLeak)
	result, err := newTestChecker(t, all.server).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, "HTTP 401")
	assertNoSecrets(t, result)
}

func TestChecker_Check_ForbiddenIsInconclusive(t *testing.T) {
	all := serve(t, http.StatusForbidden, valLeak)
	result, err := newTestChecker(t, all.server).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonProbeForbidden, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "HTTP 403")
	assertNoSecrets(t, result)
}

func TestChecker_Check_ServerErrorIsRemoteError(t *testing.T) {
	all := serve(t, http.StatusBadGateway, passBody())
	result, err := newTestChecker(t, all.server).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_RedirectIsRemoteError(t *testing.T) {
	all := serve(t, http.StatusFound, valLeak)
	result, err := newTestChecker(t, all.server).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "HTTP 302")
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

func TestChecker_Check_TlsFailure(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("the tls fixture must fail the handshake before any handler runs")
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{baseURL: server.URL}).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonTlsFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "tls handshake failed")
	assertNoSecrets(t, result)
}

func TestChecker_Check_TimeoutIsInconclusive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result, err := (&checker{baseURL: server.URL}).Check(ctx, validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_ConfigCannotOverrideEndpoint(t *testing.T) {
	fixedHits := 0
	fixed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fixedHits++
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(passBody()))
	}))
	t.Cleanup(fixed.Close)

	decoyHits := 0
	decoy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoyHits++
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(decoy.Close)

	config := validConfig()
	config["baseUrl"] = decoy.URL
	config["endpoint"] = decoy.URL
	config["apiUrl"] = decoy.URL
	config["server"] = decoy.URL
	config["ipUrl"] = decoy.URL
	config["tokenUrl"] = decoy.URL

	result, err := (&checker{baseURL: fixed.URL}).Check(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertNoSecrets(t, result)
	if decoyHits != 0 {
		t.Fatal("config-supplied endpoint keys must never be contacted")
	}
	if fixedHits != 1 {
		t.Fatalf("fixed probe hits = %d, want 1", fixedHits)
	}
}

func TestProbe_ProductionURL(t *testing.T) {
	cfg, err := parseConfig(validConfig())
	if err != nil {
		t.Fatal(err)
	}
	req, err := New().(*checker).newProbe(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if req.URL.Scheme != "https" {
		t.Fatalf("scheme = %q", req.URL.Scheme)
	}
	if req.URL.Host != "api.luadns.com" {
		t.Fatalf("host = %q", req.URL.Host)
	}
	if req.URL.Path != "/v1/zones" {
		t.Fatalf("path = %q, want /v1/zones", req.URL.Path)
	}

}



func validConfig() map[string]any {
	return map[string]any{
		"username": valUsername,
		"apiKey": valApiKey,
	}
}

func secretValues() []string {
	return []string{valUsername, valApiKey}
}

func containsSecret(text string) bool {
	for _, secret := range secretValues() {
		if secret != "" && strings.Contains(text, secret) {
			return true
		}
	}
	return false
}

func passBody() string {
	return valLeak
}

type probeRecord struct {
	method   string
	path     string
	rawQuery string
	header   http.Header
	body     string
}

type served struct {
	server *httptest.Server
	got    []probeRecord
}

func serve(t *testing.T, status int, body string) *served {
	t.Helper()
	out := &served{}
	out.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, _ := io.ReadAll(r.Body)
		out.got = append(out.got, probeRecord{
			method:   r.Method,
			path:     r.URL.Path,
			rawQuery: r.URL.RawQuery,
			header:   r.Header.Clone(),
			body:     string(payload),
		})
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(out.server.Close)
	return out
}

func newTestChecker(t *testing.T, server *httptest.Server) *checker {
	t.Helper()
	return &checker{baseURL: server.URL}
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
	if containsSecret(joined) {
		t.Errorf("diagnostics leaked secret material: %v", got.Diagnostics)
	}
}


func basicAuth(header string) (string, string, bool) {
	const prefix = "Basic "
	if !strings.HasPrefix(header, prefix) {
		return "", "", false
	}
	raw, err := base64.StdEncoding.DecodeString(header[len(prefix):])
	if err != nil {
		return "", "", false
	}
	user, passwd, ok := strings.Cut(string(raw), ":")
	return user, passwd, ok
}
