package oraclecloud

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
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
	tenancyOcid          = "ocid1.tenancy.oc1..probe"
	userOcid             = "ocid1.user.oc1..probe"
	publicKeyFingerprint = "8a:bc:probe:fingerprint"
)

var (
	privateKeyPemOnce sync.Once
	privateKeyPem     string
)

func keyPem(t *testing.T) string {
	t.Helper()
	privateKeyPemOnce.Do(func() {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return
		}
		der, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			return
		}
		privateKeyPem = string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	})
	if privateKeyPem == "" {
		t.Fatal("failed to generate a private key fixture")
	}
	return privateKeyPem
}

func TestNew_Identity(t *testing.T) {
	checker := New()
	if checker.ProviderType() != string(domain.AccessProviderTypeOracleCloud) {
		t.Fatalf("provider type = %q, want %s", checker.ProviderType(), domain.AccessProviderTypeOracleCloud)
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
		{"unsupported auth method", withField(apiKeyConfig(), "authMethod", "sessiontoken")},
		{"missing privateKey", stripField(apiKeyConfig(), "privateKey")},
		{"blank privateKey", withField(apiKeyConfig(), "privateKey", "   ")},
		{"missing publicKeyFingerprint", stripField(apiKeyConfig(), "publicKeyFingerprint")},
		{"missing tenancyOcid", stripField(apiKeyConfig(), "tenancyOcid")},
		{"missing userOcid", stripField(apiKeyConfig(), "userOcid")},
		{"instance principal missing tenancy", withField(stripField(apiKeyConfig(), "tenancyOcid"), "authMethod", "instanceprincipal")},
		{"malformed privateKey", withField(apiKeyConfig(), "privateKey", "not a pem key")},
	}

	keyPem(t)
	checker := &checker{endpoint: "http://127.0.0.1:1"}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := checker.Check(context.Background(), c.config)
			if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
				t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
			}
			if err != nil && strings.Contains(err.Error(), privateKeyPem) {
				t.Fatal("config error leaked secret material")
			}
		})
	}
}

func TestChecker_Check_Passes(t *testing.T) {
	var gotPath string
	var gotQuery string
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), apiKeyConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertDiagnostics(t, result, hostOf(server.URL), "HTTP 200", "credential accepted", "permissions on specific resources were not checked")
	assertNoSecrets(t, result)
	if gotPath != "/20180115/zones" {
		t.Fatalf("path = %q, want /20180115/zones", gotPath)
	}
	if !strings.Contains(gotQuery, "compartmentId="+tenancyOcid) || !strings.Contains(gotQuery, "limit=1") {
		t.Fatalf("query = %q, want compartmentId and limit=1", gotQuery)
	}
	if !strings.Contains(gotAuth, "Signature version=") {
		t.Fatalf("authorization header missing oci signature: %q", gotAuth)
	}
}

func TestChecker_Check_DefaultAuthMethodPasses(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusOK, `[]`))
	t.Cleanup(server.Close)

	config := apiKeyConfig()
	delete(config, "authMethod")

	result, err := checkerFor(t, server.URL).Check(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
}

func TestChecker_Check_InvalidKeyIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusUnauthorized, `{"code":"NotAuthenticated","message":"the required header x-obos-id was not supplied"}`))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), apiKeyConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, "NotAuthenticated")
	assertNoSecrets(t, result)
}

func TestChecker_Check_ForbiddenIsInconclusive(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusForbidden, `{"code":"NotAuthorizedOrNotFound","message":"not authorized"}`))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), apiKeyConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonProbeForbidden, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "NotAuthorizedOrNotFound")
	assertNoSecrets(t, result)
}

func TestChecker_Check_ServerErrorIsRemoteError(t *testing.T) {
	server := httptest.NewServer(jsonStatus(http.StatusInternalServerError, `{"code":"InternalError","message":"upstream failure"}`))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL).Check(context.Background(), apiKeyConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_RefusedConnection(t *testing.T) {
	result, err := (&checker{endpoint: "http://" + closedAddr(t)}).Check(context.Background(), apiKeyConfig())
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

	result, err := checkerFor(t, server.URL).Check(context.Background(), apiKeyConfig())
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

	result, err := checkerFor(t, server.URL).Check(ctx, apiKeyConfig())
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func apiKeyConfig() map[string]any {
	return map[string]any{
		"authMethod":           "apikey",
		"privateKey":           privateKeyPem,
		"publicKeyFingerprint": publicKeyFingerprint,
		"tenancyOcid":          tenancyOcid,
		"userOcid":             userOcid,
	}
}

func stripField(config map[string]any, field string) map[string]any {
	stripped := map[string]any{}
	for k, v := range config {
		if k != field {
			stripped[k] = v
		}
	}
	return stripped
}

func withField(config map[string]any, field string, value any) map[string]any {
	updated := map[string]any{}
	for k, v := range config {
		updated[k] = v
	}
	updated[field] = value
	return updated
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
	if strings.Contains(joined, privateKeyPem) {
		t.Fatal("diagnostics leaked secret material")
	}
}
