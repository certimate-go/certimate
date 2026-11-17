package xinnet

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

var fixedNow = time.Date(2024, 6, 5, 8, 55, 45, 0, time.UTC)

const (
	valLeak        = "probe-leak-xinnet"
	valAgentId     = "probe-xinnet-agentId"
	valApiPassword = "probe-xinnet-apiPassword"
)

func TestNew_Identity(t *testing.T) {
	checker := New()
	if checker.ProviderType() != string(domain.AccessProviderTypeXinnet) {
		t.Fatalf("provider type = %q, want %s", checker.ProviderType(), domain.AccessProviderTypeXinnet)
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
		{"missing agentId", map[string]any{"apiPassword": valApiPassword}},
		{"blank agentId", map[string]any{"agentId": "   ", "apiPassword": valApiPassword}},
		{"malformed agentId", map[string]any{"agentId": 1, "apiPassword": valApiPassword}},
		{"missing apiPassword", map[string]any{"agentId": valAgentId}},
		{"blank apiPassword", map[string]any{"apiPassword": "   ", "agentId": valAgentId}},
		{"malformed apiPassword", map[string]any{"apiPassword": 1, "agentId": valAgentId}},
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
	if got.method != http.MethodPost {
		t.Fatalf("method = %q", got.method)
	}
	if got.path != "/api/domain/list" {
		t.Fatalf("path = %q, want /api/domain/list", got.path)
	}
	if strings.Contains(got.body, valApiPassword) {
		t.Fatal("api password was sent in the body")
	}
	if got.header.Get("Timestamp") != fixedNow.UTC().Format("20060102T150405Z") {
		t.Fatalf("timestamp = %q", got.header.Get("Timestamp"))
	}
	sig := signXinnet(valApiPassword, got.header.Get("Timestamp"), http.MethodPost, got.path+"/", got.body)
	if got.header.Get("Authorization") != "HMAC-SHA256 Access="+valAgentId+", Signature="+sig {
		t.Fatalf("authorization = %q", got.header.Get("Authorization"))
	}
}

func TestChecker_Check_BadSignatureIsAuthRejected(t *testing.T) {
	all := serve(t, http.StatusOK, `{"code":"350002","msg":"`+valApiPassword+`"}`)
	result, err := newTestChecker(t, all.server).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, "code 350002")
	assertNoSecrets(t, result)
}

func TestChecker_Check_IPDeniedIsForbidden(t *testing.T) {
	all := serve(t, http.StatusOK, `{"code":"355020","msg":"`+valApiPassword+`"}`)
	result, err := newTestChecker(t, all.server).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonProbeForbidden, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "code 355020")
	assertNoSecrets(t, result)
}

func TestChecker_Check_OtherCodeIsRemote(t *testing.T) {
	all := serve(t, http.StatusOK, `{"code":"355050","msg":"`+valApiPassword+`"}`)
	result, err := newTestChecker(t, all.server).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "code 355050")
	assertNoSecrets(t, result)
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
	if req.URL.Host != "apiv2.xinnet.com" {
		t.Fatalf("host = %q", req.URL.Host)
	}
	if req.URL.Path != "/api/domain/list" {
		t.Fatalf("path = %q, want /api/domain/list", req.URL.Path)
	}
}

func TestSignXinnet_DocVector(t *testing.T) {
	got := signXinnet("9f3b6760952b41eda7ddad84d755b3944cbb7929277947c294b7a578128e1170", "20240605T085545Z", "POST", "/api/domain/check/", "{}")
	if got != "bbcebdf1d93fbd8dfbd18313b29e645014c98e307d8f3a88b7517199bad564a6" {
		t.Fatalf("signature = %s", got)
	}
}

func validConfig() map[string]any {
	return map[string]any{
		"agentId":     valAgentId,
		"apiPassword": valApiPassword,
	}
}

func secretValues() []string {
	return []string{valAgentId, valApiPassword}
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
	return `{"code":"0","msg":"ok"}`
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
	return &checker{baseURL: server.URL, now: func() time.Time { return fixedNow }}
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
