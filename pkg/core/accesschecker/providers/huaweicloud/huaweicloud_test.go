package huaweicloud

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
)

const (
	accessKeyId     = "HWprobeAccessKeyIdValue"
	secretAccessKey = "probe-huawei-secret-value"
)

func TestNew_Identity(t *testing.T) {
	checker := New()
	if checker.ProviderType() != string(domain.AccessProviderTypeHuaweiCloud) {
		t.Fatalf("provider type = %q, want %s", checker.ProviderType(), domain.AccessProviderTypeHuaweiCloud)
	}
	scope := checker.Scope()
	if len(scope) != 2 || scope[0] != core.AccessCheckItemConnect || scope[1] != core.AccessCheckItemAuth {
		t.Fatalf("scope = %v, want connect and auth", scope)
	}
}

func TestChecker_Check_ConfigInvalid(t *testing.T) {
	cases := []map[string]any{
		nil,
		{},
		{"secretAccessKey": secretAccessKey},
		{"accessKeyId": "  ", "secretAccessKey": secretAccessKey},
		{"accessKeyId": accessKeyId},
		{"accessKeyId": accessKeyId, "secretAccessKey": "  "},
		{"accessKeyId": true, "secretAccessKey": secretAccessKey},
	}
	checker := &checker{endpoint: "http://127.0.0.1:1"}
	for _, config := range cases {
		_, err := checker.Check(context.Background(), config)
		if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
			t.Fatalf("config %v: expected ErrAccessCheckConfigInvalid, got %v", config, err)
		}
		if err != nil && (strings.Contains(err.Error(), accessKeyId) || strings.Contains(err.Error(), secretAccessKey)) {
			t.Fatal("config error leaked secret material")
		}
	}
}

func TestChecker_Check_Passes(t *testing.T) {
	var path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"zones":[]}`))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{endpoint: server.URL}).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertNoSecrets(t, result)
	if !strings.Contains(path, "zone") {
		t.Fatalf("path = %q, want a public zone list", path)
	}
}

func TestChecker_Check_InvalidKeyIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error_code":"APIGW.0301","error_msg":"Incorrect IAM authentication information","request_id":"req-1"}`))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{endpoint: server.URL}).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertNoSecrets(t, result)
}

func TestChecker_Check_RefusedConnection(t *testing.T) {
	result, err := (&checker{endpoint: "http://" + closedAddr(t)}).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonConnectFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_TlsFailure(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("the tls fixture must fail the handshake before any handler runs")
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{endpoint: server.URL}).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonTlsFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_TimeoutIsInconclusive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result, err := (&checker{endpoint: server.URL}).Check(ctx, validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func validConfig() map[string]any {
	return map[string]any{"accessKeyId": accessKeyId, "secretAccessKey": secretAccessKey}
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

func assertResult(t *testing.T, got *core.AccessCheckResult, wantStatus, wantReason, wantConnect, wantAuth string) {
	t.Helper()
	if got == nil {
		t.Fatal("expected a result, got nil")
	}
	if got.Status != wantStatus || got.ReasonCode != wantReason {
		t.Fatalf("status/reason = %s/%s, want %s/%s diagnostics=%v", got.Status, got.ReasonCode, wantStatus, wantReason, got.Diagnostics)
	}
	if len(got.Items) != 2 || got.Items[0].Status != wantConnect || got.Items[1].Status != wantAuth {
		t.Fatalf("items = %+v", got.Items)
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
