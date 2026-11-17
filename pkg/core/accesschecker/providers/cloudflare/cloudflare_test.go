package cloudflare

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/certimate-go/certimate/pkg/core"
)

const (
	activeBody       = `{"result":{"id":"9a7806061c88ada191ed06f989cc3d26","status":"active"},"success":true,"errors":[],"messages":[]}`
	expiredBody      = `{"result":{"id":"9a7806061c88ada191ed06f989cc3d26","status":"expired"},"success":true,"errors":[],"messages":[]}`
	pendingBody      = `{"result":{"id":"9a7806061c88ada191ed06f989cc3d26","status":"pending"},"success":true,"errors":[],"messages":[]}`
	invalidTokenBody = `{"success":false,"errors":[{"code":1000,"message":"Invalid API Tokens"}],"messages":[],"result":null}`
	badHeaderBody    = `{"success":false,"errors":[{"code":6003,"message":"Invalid request headers","error_chain":[{"code":6111,"message":"Invalid format for Authorization header"}]}],"messages":[],"result":null}`
	forbiddenBody    = `{"success":false,"errors":[{"code":9109,"message":"Unauthorized to access requested resource"}],"messages":[],"result":null}`
	rateLimitBody    = `{"success":false,"errors":[{"code":9715,"message":"You have exceeded the rate limit for this endpoint"}],"messages":[],"result":null}`
)

func TestChecker_Check_ConfigInvalid(t *testing.T) {
	cases := []struct {
		name   string
		config map[string]any
	}{
		{"empty config", map[string]any{}},
		{"missing apiToken", map[string]any{"apiTokenForZone": "tok-zone"}},
		{"blank apiToken", map[string]any{"apiToken": "   "}},
		{"malformed apiToken type", map[string]any{"apiToken": 12345}},
		{"malformed apiTokenForZone type", map[string]any{"apiToken": "tok-main", "apiTokenForZone": true}},
	}

	checker := &checker{baseURL: "http://127.0.0.1:1"}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := checker.Check(context.Background(), c.config)
			if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
				t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
			}
		})
	}
}

func TestChecker_Check_ActiveTokensPass(t *testing.T) {
	var mu sync.Mutex
	requests := []probeRecord{}
	server := httptest.NewServer(verifyFixture(&mu, &requests, func(auth string) (int, string) {
		return http.StatusOK, activeBody
	}))
	t.Cleanup(server.Close)

	checker := &checker{baseURL: server.URL}
	config := map[string]any{"apiToken": "tok-main", "apiTokenForZone": "tok-zone"}

	result, err := checker.Check(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertDiagnostics(t, result, hostOf(server.URL), "token apiToken: token is active", "token apiTokenForZone: token is active", "permissions on specific resources were not checked")

	if len(requests) != 2 {
		t.Fatalf("expected one probe per filled token, got %d", len(requests))
	}
	for _, req := range requests {
		if req.path != "/user/tokens/verify" {
			t.Errorf("probe path = %q, want /user/tokens/verify", req.path)
		}
		if req.auth != "Bearer tok-main" && req.auth != "Bearer tok-zone" {
			t.Errorf("probe authorization = %q, want a bearer token from the config", req.auth)
		}
	}
}

func TestChecker_Check_SingleTokenProbesOnce(t *testing.T) {
	var mu sync.Mutex
	requests := []probeRecord{}
	server := httptest.NewServer(verifyFixture(&mu, &requests, func(auth string) (int, string) {
		return http.StatusOK, activeBody
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{baseURL: server.URL}).Check(context.Background(), map[string]any{"apiToken": "tok-main"})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	if len(requests) != 1 {
		t.Fatalf("expected a single probe for apiToken only, got %d", len(requests))
	}
	if requests[0].auth != "Bearer tok-main" {
		t.Fatalf("probe authorization = %q, want Bearer tok-main", requests[0].auth)
	}
}

func TestChecker_Check_InvalidZoneTokenIsAuthRejected(t *testing.T) {
	var mu sync.Mutex
	requests := []probeRecord{}
	server := httptest.NewServer(verifyFixture(&mu, &requests, func(auth string) (int, string) {
		if auth == "Bearer tok-zone" {
			return http.StatusBadRequest, invalidTokenBody
		}
		return http.StatusOK, activeBody
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{baseURL: server.URL}).Check(context.Background(), map[string]any{"apiToken": "tok-main", "apiTokenForZone": "tok-zone"})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, "auth_rejected", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, hostOf(server.URL), "token apiTokenForZone: api reported the token as invalid")
	if len(requests) != 2 {
		t.Fatalf("expected both tokens probed, got %d", len(requests))
	}
}

func TestChecker_Check_ExpiredTokenStatusIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expiredBody))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{baseURL: server.URL}).Check(context.Background(), map[string]any{"apiToken": "tok-main"})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, "auth_rejected", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, hostOf(server.URL), "token apiToken: token status is expired")
}

func TestChecker_Check_BadHeaderTokenIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(badHeaderBody))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{baseURL: server.URL}).Check(context.Background(), map[string]any{"apiToken": "tok-main", "apiTokenForZone": "tok-zone"})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, "auth_rejected", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, hostOf(server.URL), "token apiToken: api reported the token as invalid (6003 Invalid request headers: 6111 Invalid format for Authorization header)")
}

func TestChecker_Check_ForbiddenIsInconclusive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(forbiddenBody))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{baseURL: server.URL}).Check(context.Background(), map[string]any{"apiToken": "tok-main"})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, "probe_forbidden", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, hostOf(server.URL), "token apiToken: endpoint returned HTTP 403")
}

func TestChecker_Check_RateLimitIsInconclusive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(rateLimitBody))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{baseURL: server.URL}).Check(context.Background(), map[string]any{"apiToken": "tok-main"})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, "remote_error", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
}

func TestChecker_Check_ServerErrorIsInconclusive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("upstream panicked"))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{baseURL: server.URL}).Check(context.Background(), map[string]any{"apiToken": "tok-main"})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, "remote_error", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
}

func TestChecker_Check_UnknownTokenStatusIsInconclusive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(pendingBody))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{baseURL: server.URL}).Check(context.Background(), map[string]any{"apiToken": "tok-main"})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, "remote_error", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, hostOf(server.URL), "token apiToken: no credential verdict")
}

func TestChecker_Check_MixedPassAndRemoteErrorIsInconclusive(t *testing.T) {
	server := httptest.NewServer(verifyFixture(nil, nil, func(auth string) (int, string) {
		if auth == "Bearer tok-zone" {
			return http.StatusInternalServerError, "upstream panicked"
		}
		return http.StatusOK, activeBody
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{baseURL: server.URL}).Check(context.Background(), map[string]any{"apiToken": "tok-main", "apiTokenForZone": "tok-zone"})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, "remote_error", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
}

func TestChecker_Check_TimeoutIsInconclusive(t *testing.T) {
	server := httptest.NewServer(hangingFixture())
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result, err := (&checker{baseURL: server.URL}).Check(ctx, map[string]any{"apiToken": "tok-main"})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, "timeout", core.AccessCheckItemStatusSkipped, core.AccessCheckItemStatusSkipped)
}

func TestChecker_Check_CancelIsInconclusive(t *testing.T) {
	server := httptest.NewServer(hangingFixture())
	t.Cleanup(server.Close)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	result, err := (&checker{baseURL: server.URL}).Check(ctx, map[string]any{"apiToken": "tok-main"})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, "canceled", core.AccessCheckItemStatusSkipped, core.AccessCheckItemStatusSkipped)
}

func TestChecker_Check_TlsFailureIsConnectFailed(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("the tls fixture must fail the handshake before any handler runs")
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{baseURL: server.URL}).Check(context.Background(), map[string]any{"apiToken": "tok-main"})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, "tls_failed", core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, hostOf(server.URL), "tls handshake failed")
}

func TestChecker_Check_RefusedConnectionIsConnectFailed(t *testing.T) {
	base := "http://" + closedAddr(t)

	result, err := (&checker{baseURL: base}).Check(context.Background(), map[string]any{"apiToken": "tok-main"})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, "connect_failed", core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, hostOf(base), "tcp or dns failure")
}

func TestChecker_Check_ConfigCannotOverrideEndpoint(t *testing.T) {
	var mu sync.Mutex
	requests := []probeRecord{}
	fixed := httptest.NewServer(verifyFixture(&mu, &requests, func(auth string) (int, string) {
		return http.StatusOK, activeBody
	}))
	t.Cleanup(fixed.Close)

	decoyContacted := false
	decoy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoyContacted = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(activeBody))
	}))
	t.Cleanup(decoy.Close)

	config := map[string]any{
		"apiToken":        "tok-main",
		"baseUrl":         decoy.URL,
		"baseURL":         decoy.URL,
		"endpoint":        decoy.URL,
		"apiUrl":          decoy.URL,
		"serverUrl":       decoy.URL,
		"apiTokenForZone": "tok-zone",
	}

	result, err := (&checker{baseURL: fixed.URL}).Check(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	if decoyContacted {
		t.Fatal("config-supplied endpoint keys must never be contacted")
	}
	if len(requests) != 2 {
		t.Fatalf("expected both probes against the fixed base url, got %d", len(requests))
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

type probeRecord struct {
	path string
	auth string
}

func verifyFixture(mu *sync.Mutex, requests *[]probeRecord, respond func(auth string) (int, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if mu != nil && requests != nil {
			mu.Lock()
			*requests = append(*requests, probeRecord{path: r.URL.Path, auth: auth})
			mu.Unlock()
		}
		code, body := respond(auth)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		w.Write([]byte(body))
	}
}

func hangingFixture() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
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

func assertDiagnostics(t *testing.T, got *core.AccessCheckResult, host string, mustContain ...string) {
	t.Helper()

	joined := strings.Join(got.Diagnostics, "\n")
	if !strings.Contains(joined, host) {
		t.Errorf("diagnostics missing endpoint host %q: %v", host, got.Diagnostics)
	}
	for _, substr := range mustContain {
		if !strings.Contains(joined, substr) {
			t.Errorf("diagnostics missing %q: %v", substr, got.Diagnostics)
		}
	}
	for _, secret := range []string{"tok-main", "tok-zone"} {
		if strings.Contains(joined, secret) {
			t.Errorf("diagnostics leaked token material %q: %v", secret, got.Diagnostics)
		}
	}
}
