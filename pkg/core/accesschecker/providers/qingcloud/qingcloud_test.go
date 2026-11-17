package qingcloud

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
)

const (
	accessKeyId     = "QINGprobeAccessKeyId"
	secretAccessKey = "probe-qingcloud-secret-value"
)

func TestNew_Identity(t *testing.T) {
	checker := New()
	if checker.ProviderType() != string(domain.AccessProviderTypeQingCloud) {
		t.Fatalf("provider type = %q, want %s", checker.ProviderType(), domain.AccessProviderTypeQingCloud)
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
		{"missing accessKeyId", map[string]any{"secretAccessKey": secretAccessKey}},
		{"blank accessKeyId", map[string]any{"accessKeyId": "   ", "secretAccessKey": secretAccessKey}},
		{"missing secretAccessKey", map[string]any{"accessKeyId": accessKeyId}},
		{"blank secretAccessKey", map[string]any{"accessKeyId": accessKeyId, "secretAccessKey": "   "}},
		{"malformed accessKeyId", map[string]any{"accessKeyId": 1, "secretAccessKey": secretAccessKey}},
	}

	checker := &checker{endpoint: "http://127.0.0.1:1"}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := checker.Check(context.Background(), c.config)
			if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
				t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
			}
			if err != nil && (strings.Contains(err.Error(), secretAccessKey) || strings.Contains(err.Error(), accessKeyId)) {
				t.Fatal("config error leaked secret material")
			}
		})
	}
}

func signedHandler(t *testing.T, status int, body string, onSeen func(r *http.Request)) http.HandlerFunc {
	t.Helper()
	var mu sync.Mutex
	return func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		date := r.Header.Get("Date")
		authorization := r.Header.Get("Authorization")
		parts := strings.SplitN(authorization, " ", 2)
		if len(parts) != 2 || parts[0] != "QC-HMAC-SHA256" {
			t.Errorf("authorization scheme = %q", authorization)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		keyParts := strings.SplitN(parts[1], ":", 2)
		if len(keyParts) != 2 || keyParts[0] != accessKeyId {
			t.Errorf("authorization access key = %q", parts[1])
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		canonicalResource := r.URL.Path + "?" + r.URL.Query().Encode()
		mac := hmac.New(sha256.New, []byte(secretAccessKey))
		mac.Write([]byte(r.Method + "\n" + date + "\n" + canonicalResource))
		expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
		if keyParts[1] != expected {
			t.Errorf("signature mismatch: got %q want %q", keyParts[1], expected)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if onSeen != nil {
			onSeen(r)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = fmt.Fprint(w, body)
	}
}

func TestChecker_Check_Passes(t *testing.T) {
	var gotPath string
	var gotQuery string
	server := httptest.NewServer(signedHandler(t, http.StatusOK, `{"code":0,"message":"","total_count":0,"zones":[]}`, func(r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
	}))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertDiagnostics(t, result, hostOf(server.URL), "HTTP 200", "credential accepted", "permissions on specific resources were not checked")
	assertNoSecrets(t, result)
	if gotPath != "/v1/user/zones" {
		t.Fatalf("path = %q, want /v1/user/zones", gotPath)
	}
	if !strings.Contains(gotQuery, "limit=1") {
		t.Fatalf("query = %q, want limit=1", gotQuery)
	}
}

func TestChecker_Check_InvalidKeyIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusUnauthorized, `{"code":1400,"message":"signature is not valid"}`))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, "signature is not valid")
	assertNoSecrets(t, result)
}

func TestChecker_Check_ForbiddenIsInconclusive(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusForbidden, `{"code":1403,"message":"permission denied"}`))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonProbeForbidden, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_ServerErrorIsRemoteError(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusBadGateway, `{"code":1502,"message":"internal error"}`))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_BodyFaultIsRemoteError(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusOK, `{"code":1300,"message":"account locked"}`))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "account locked")
	assertNoSecrets(t, result)
}

func TestChecker_Check_RefusedConnection(t *testing.T) {
	result, err := (&checker{endpoint: "http://" + closedAddr(t)}).Check(context.Background(), validConfig())
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

	result, err := checkerFor(t, server.URL).Check(context.Background(), validConfig())
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

	result, err := checkerFor(t, server.URL).Check(ctx, validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_ConfigCannotOverrideEndpoint(t *testing.T) {
	fixed := httptest.NewServer(jsonStatus(http.StatusOK, `{"code":0,"message":"","total_count":0,"zones":[]}`))
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

	result, err := checkerFor(t, fixed.URL).Check(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	if decoyHits != 0 {
		t.Fatalf("config-supplied endpoint was contacted %d times", decoyHits)
	}
}

func validConfig() map[string]any {
	return map[string]any{
		"accessKeyId":     accessKeyId,
		"secretAccessKey": secretAccessKey,
	}
}

func checkerFor(t *testing.T, raw string) *checker {
	t.Helper()
	return &checker{endpoint: raw}
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
	for _, secret := range []string{accessKeyId, secretAccessKey} {
		if strings.Contains(joined, secret) {
			t.Fatal("diagnostics leaked secret material")
		}
	}
}
