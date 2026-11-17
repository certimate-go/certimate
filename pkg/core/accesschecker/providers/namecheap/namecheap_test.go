package namecheap

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

const (
	valLeak     = "probe-leak-namecheap"
	valUsername = "probe-namecheap-username"
	valApiKey   = "probe-namecheap-apiKey"
	discovered  = "203.0.113.8"
)

func TestNew_Identity(t *testing.T) {
	checker := New()
	if checker.ProviderType() != string(domain.AccessProviderTypeNamecheap) {
		t.Fatalf("provider type = %q, want %s", checker.ProviderType(), domain.AccessProviderTypeNamecheap)
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
		{"blank apiKey", map[string]any{"username": valUsername, "apiKey": "   "}},
		{"malformed apiKey", map[string]any{"username": valUsername, "apiKey": 1}},
	}

	checker := &checker{apiURL: "http://127.0.0.1:1/xml.response", ipURL: "http://127.0.0.1:1/getip"}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := checker.Check(context.Background(), c.config)
			if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
				t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
			}
			if err != nil && (strings.Contains(err.Error(), valApiKey) || strings.Contains(err.Error(), valUsername)) {
				t.Fatalf("config error leaked secret material: %v", err)
			}
		})
	}
}

func TestChecker_Check_Passes(t *testing.T) {
	all := serveAPI(t, http.StatusOK, successXML())
	cfg := validConfig()
	cfg["clientIp"] = "198.51.100.9"
	result, err := checkerFor(all.server).Check(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertDiagnostics(t, result, hostOf(all.server.URL), "HTTP 200", "credential accepted")
	assertNoSecrets(t, result)
	if len(all.got) != 2 {
		t.Fatalf("probe hits = %d, want 2", len(all.got))
	}
	if !strings.HasSuffix(all.got[0].path, "/getip") {
		t.Fatalf("ip path = %q", all.got[0].path)
	}
	got := all.got[1]
	if got.method != http.MethodGet || !strings.HasSuffix(got.path, "/xml.response") {
		t.Fatalf("api request %s %s", got.method, got.path)
	}
	q, err := url.ParseQuery(got.rawQuery)
	if err != nil {
		t.Fatal(err)
	}
	if q.Get("ApiUser") != valUsername || q.Get("UserName") != valUsername || q.Get("ApiKey") != valApiKey || q.Get("Command") != "namecheap.domains.getList" || q.Get("ClientIp") != discovered || q.Get("Page") != "1" || q.Get("PageSize") != "10" {
		t.Fatalf("query = %v", q)
	}
}

func TestChecker_Check_InvalidKeyIsAuthRejected(t *testing.T) {
	all := serveAPI(t, http.StatusOK, errorXML("1011102"))
	result, err := checkerFor(all.server).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, "error 1011102")
	assertNoSecrets(t, result)
}

func TestChecker_Check_InvalidIPIsForbidden(t *testing.T) {
	all := serveAPI(t, http.StatusOK, errorXML("1011150"))
	result, err := checkerFor(all.server).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonProbeForbidden, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "error 1011150")
	assertNoSecrets(t, result)
}

func TestChecker_Check_OtherXMLErrorIsRemote(t *testing.T) {
	all := serveAPI(t, http.StatusOK, errorXML("9999"))
	result, err := checkerFor(all.server).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "error 9999")
	assertNoSecrets(t, result)
}

func TestChecker_Check_UnauthorizedIsAuthRejected(t *testing.T) {
	all := serveAPI(t, http.StatusUnauthorized, valApiKey)
	result, err := checkerFor(all.server).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, "HTTP 401")
	assertNoSecrets(t, result)
}

func TestChecker_Check_ForbiddenIsInconclusive(t *testing.T) {
	all := serveAPI(t, http.StatusForbidden, valApiKey)
	result, err := checkerFor(all.server).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonProbeForbidden, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_ServerErrorIsRemoteError(t *testing.T) {
	all := serveAPI(t, http.StatusBadGateway, successXML())
	result, err := checkerFor(all.server).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_RedirectIsRemoteError(t *testing.T) {
	all := serveAPI(t, http.StatusFound, valApiKey)
	result, err := checkerFor(all.server).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "HTTP 302")
	assertNoSecrets(t, result)
}

func TestChecker_Check_RefusedConnection(t *testing.T) {
	addr := "http://" + closedAddr(t)
	result, err := (&checker{apiURL: addr + "/xml.response", ipURL: addr + "/getip"}).Check(context.Background(), validConfig())
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

	result, err := checkerFor(server).Check(context.Background(), validConfig())
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

	result, err := checkerFor(server).Check(ctx, validConfig())
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
		if strings.HasSuffix(r.URL.Path, "/getip") {
			_, _ = w.Write([]byte(discovered))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(successXML()))
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
	config["ipUrl"] = decoy.URL
	config["server"] = decoy.URL

	result, err := checkerFor(fixed).Check(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertNoSecrets(t, result)
	if decoyHits != 0 {
		t.Fatal("config-supplied endpoint keys must never be contacted")
	}
	if fixedHits != 2 {
		t.Fatalf("fixed probe hits = %d, want 2", fixedHits)
	}
}

func TestProbe_ProductionURL(t *testing.T) {
	checker := New().(*checker)
	if checker.apiURL != defaultAPIURL {
		t.Fatalf("api url = %q", checker.apiURL)
	}
	if checker.ipURL != defaultIPURL {
		t.Fatalf("ip url = %q", checker.ipURL)
	}
}

func validConfig() map[string]any {
	return map[string]any{
		"username": valUsername,
		"apiKey":   valApiKey,
	}
}

func successXML() string {
	return `<ApiResponse Status="OK" xmlns="http://api.namecheap.com/xml.response"><Errors /><CommandResponse Type="namecheap.domains.getList"></CommandResponse></ApiResponse>`
}

func errorXML(number string) string {
	return `<ApiResponse Status="ERROR" xmlns="http://api.namecheap.com/xml.response"><Errors><Error Number="` + number + `">` + valApiKey + `</Error></Errors></ApiResponse>`
}

type probeRecord struct {
	method   string
	path     string
	rawQuery string
}

type served struct {
	server *httptest.Server
	got    []probeRecord
}

func serveAPI(t *testing.T, status int, body string) *served {
	t.Helper()
	out := &served{}
	out.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		out.got = append(out.got, probeRecord{method: r.Method, path: r.URL.Path, rawQuery: r.URL.RawQuery})
		if strings.HasSuffix(r.URL.Path, "/getip") {
			_, _ = w.Write([]byte(discovered + "\n"))
			return
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(out.server.Close)
	return out
}

func checkerFor(server *httptest.Server) *checker {
	return &checker{apiURL: server.URL + "/xml.response", ipURL: server.URL + "/getip"}
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
	if strings.Contains(joined, valApiKey) || strings.Contains(joined, valUsername) || strings.Contains(joined, valLeak) {
		t.Errorf("diagnostics leaked secret material: %v", got.Diagnostics)
	}
}
