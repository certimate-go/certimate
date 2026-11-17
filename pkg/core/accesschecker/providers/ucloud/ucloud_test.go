package ucloud

import (
	"context"
	"errors"
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
	publicKey  = "ucprobePublicKeyToken"
	privateKey = "probe-ucloud-private-value"
)

func TestNew_Identity(t *testing.T) {
	checker := New()
	if checker.ProviderType() != string(domain.AccessProviderTypeUCloud) {
		t.Fatalf("provider type = %q, want %s", checker.ProviderType(), domain.AccessProviderTypeUCloud)
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
		{"missing publicKey", map[string]any{"privateKey": privateKey}},
		{"blank publicKey", map[string]any{"publicKey": "   ", "privateKey": privateKey}},
		{"missing privateKey", map[string]any{"publicKey": publicKey}},
		{"blank privateKey", map[string]any{"publicKey": publicKey, "privateKey": "   "}},
		{"malformed publicKey", map[string]any{"publicKey": 1, "privateKey": privateKey}},
	}

	checker := &checker{endpoint: "http://127.0.0.1:1"}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := checker.Check(context.Background(), c.config)
			if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
				t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
			}
			if err != nil && (strings.Contains(err.Error(), privateKey) || strings.Contains(err.Error(), publicKey)) {
				t.Fatal("config error leaked secret material")
			}
		})
	}
}

func TestChecker_Check_Passes(t *testing.T) {
	var gotAction string
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, 1<<16)
		n, _ := r.Body.Read(body)
		gotAction = r.URL.Query().Get("Action")
		gotBody = string(body[:n])
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"RetCode":0,"Message":"","TotalCount":1,"ProjectIds":["org-project"]}`))
	}))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertDiagnostics(t, result, hostOf(server.URL), "HTTP 200", "credential accepted", "permissions on specific resources were not checked")
	assertNoSecrets(t, result)
	if gotAction != "GetProjectList" {
		t.Fatalf("action = %q, want GetProjectList", gotAction)
	}
	if !strings.Contains(gotBody, "Action=GetProjectList") || !strings.Contains(gotBody, "PublicKey="+publicKey) {
		t.Fatalf("probe body missing action or public key: %s", gotBody)
	}
	if strings.Contains(gotBody, privateKey) {
		t.Fatal("probe sent the private key in the clear")
	}
}

func TestChecker_Check_InvalidKeyIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusOK, `{"RetCode":230,"Message":"public/private key pair is invalid"}`))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, "public/private key pair is invalid")
	assertNoSecrets(t, result)
}

func TestChecker_Check_SignatureFaultIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusOK, `{"RetCode":231,"Message":"signature verification failed"}`))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertNoSecrets(t, result)
}

func TestChecker_Check_HttpUnauthorizedIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusUnauthorized, `{"RetCode":-1,"Message":"unauthorized"}`))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertNoSecrets(t, result)
}

func TestChecker_Check_ForbiddenIsInconclusive(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusForbidden, `{"RetCode":-1,"Message":"forbidden"}`))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonProbeForbidden, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_RetCodeFaultIsRemoteError(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusOK, `{"RetCode":2300,"Message":"project list unavailable"}`))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "project list unavailable")
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
		timer := time.NewTimer(time.Second)
		defer timer.Stop()
		select {
		case <-r.Context().Done():
		case <-timer.C:
		}
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
	fixed := httptest.NewServer(jsonStatus(http.StatusOK, `{"RetCode":0,"Message":"","TotalCount":0}`))
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
		"publicKey":  publicKey,
		"privateKey": privateKey,
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
	for _, secret := range []string{publicKey, privateKey} {
		if strings.Contains(joined, secret) {
			t.Fatal("diagnostics leaked secret material")
		}
	}
}
