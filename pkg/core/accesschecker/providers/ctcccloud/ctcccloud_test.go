package ctcccloud

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
	accessKeyId     = "ctccprobeAccessKeyId"
	secretAccessKey = "probe-ctcc-secret-value"
)

func TestNew_Identity(t *testing.T) {
	checker := New()
	if checker.ProviderType() != string(domain.AccessProviderTypeCTCCCloud) {
		t.Fatalf("provider type = %q, want %s", checker.ProviderType(), domain.AccessProviderTypeCTCCCloud)
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

func TestChecker_Check_Passes(t *testing.T) {
	var gotPath string
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"statusCode":"200","message":"ok","returnObj":{"records":[]}}`))
	}))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertDiagnostics(t, result, hostOf(server.URL), "HTTP 200", "credential accepted", "permissions on specific resources were not checked")
	assertNoSecrets(t, result)
	if gotPath != "/v2/queryRecordList" {
		t.Fatalf("path = %q, want /v2/queryRecordList", gotPath)
	}
	if !strings.Contains(gotQuery, "domain="+probeDomain) {
		t.Fatalf("query = %q, want domain=%s", gotQuery, probeDomain)
	}
	if strings.Contains(gotQuery, secretAccessKey) {
		t.Fatal("probe sent the access key secret in the clear")
	}
}

func TestChecker_Check_InvalidKeyIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusBadRequest, `{"statusCode":"5001003","errorMessage":"invalid access key id"}`))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, "invalid access key id")
	assertNoSecrets(t, result)
}

func TestChecker_Check_BodyAuthFaultIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusOK, `{"statusCode":"5001004","errorMessage":"signature does not match"}`))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, "signature does not match")
	assertNoSecrets(t, result)
}

func TestChecker_Check_ForbiddenIsInconclusive(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusForbidden, `{"statusCode":"403","errorMessage":"no permission"}`))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonProbeForbidden, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_ServerErrorIsRemoteError(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusBadGateway, `{"statusCode":"502","errorMessage":"upstream"}`))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
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
	fixed := httptest.NewServer(jsonStatus(http.StatusOK, `{"statusCode":"200","message":"ok"}`))
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
