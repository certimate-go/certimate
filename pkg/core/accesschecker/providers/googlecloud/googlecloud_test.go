package googlecloud

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
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
	projectId   = "probe-project"
	clientEmail = "probe@probe-project.iam.gserviceaccount.com"
	accessToken = "probe-google-access-token"
)

func TestNew_Identity(t *testing.T) {
	checker := New()
	if checker.ProviderType() != string(domain.AccessProviderTypeGoogleCloud) {
		t.Fatalf("provider type = %q, want %s", checker.ProviderType(), domain.AccessProviderTypeGoogleCloud)
	}
	scope := checker.Scope()
	if len(scope) != 2 || scope[0] != core.AccessCheckItemConnect || scope[1] != core.AccessCheckItemAuth {
		t.Fatalf("scope = %v, want connect and auth", scope)
	}
}

func TestChecker_Check_ConfigInvalid(t *testing.T) {
	key := mustKey(t)
	cases := []map[string]any{
		nil,
		{},
		{"projectId": projectId},
		{"projectId": " ", "serviceAccountKey": key},
		{"projectId": "../etc", "serviceAccountKey": key},
		{"projectId": projectId, "serviceAccountKey": "not-json"},
		{"projectId": projectId, "serviceAccountKey": `{"type":"service_account"}`},
	}
	checker := &checker{tokenURL: "http://127.0.0.1:1/token", dnsURL: "http://127.0.0.1:1"}
	for _, config := range cases {
		_, err := checker.Check(context.Background(), config)
		if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
			t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
		}
		if err != nil && strings.Contains(err.Error(), "PRIVATE KEY") {
			t.Fatal("config error leaked secret material")
		}
	}
}

func TestChecker_Check_Passes(t *testing.T) {
	key := mustKey(t)
	var sawToken, sawDNS bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/token":
			sawToken = true
			if err := r.ParseForm(); err != nil {
				t.Error(err)
			}
			if r.Form.Get("grant_type") != "urn:ietf:params:oauth:grant-type:jwt-bearer" || r.Form.Get("assertion") == "" {
				t.Fatal("token request was not a jwt bearer grant")
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"` + accessToken + `","token_type":"Bearer"}`))
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/dns/v1/projects/"+projectId+"/managedZones"):
			sawDNS = true
			if r.Header.Get("Authorization") != "Bearer "+accessToken {
				t.Error("dns request did not use the acquired token")
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"managedZones":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{tokenURL: server.URL + "/token", dnsURL: server.URL}).Check(context.Background(), validConfig(key))
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertNoSecrets(t, result, key)
	if !sawToken || !sawDNS {
		t.Fatalf("token=%v dns=%v, want both probes", sawToken, sawDNS)
	}
}

func TestChecker_Check_InvalidGrantIsAuthRejected(t *testing.T) {
	key := mustKey(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"bad key"}`))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{tokenURL: server.URL + "/token", dnsURL: server.URL}).Check(context.Background(), validConfig(key))
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertNoSecrets(t, result, key)
}

func TestChecker_Check_DnsForbiddenStillPasses(t *testing.T) {
	key := mustKey(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"` + accessToken + `"}`))
			return
		}
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"code":403,"message":"forbidden"}}`))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{tokenURL: server.URL + "/token", dnsURL: server.URL}).Check(context.Background(), validConfig(key))
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertNoSecrets(t, result, key)
}

func TestChecker_Check_RefusedConnection(t *testing.T) {
	key := mustKey(t)
	base := "http://" + closedAddr(t)
	result, err := (&checker{tokenURL: base + "/token", dnsURL: base}).Check(context.Background(), validConfig(key))
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonConnectFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result, key)
}

func TestChecker_Check_TlsFailure(t *testing.T) {
	key := mustKey(t)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("the tls fixture must fail the handshake before any handler runs")
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{tokenURL: server.URL + "/token", dnsURL: server.URL}).Check(context.Background(), validConfig(key))
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonTlsFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result, key)
}

func TestChecker_Check_TimeoutIsInconclusive(t *testing.T) {
	key := mustKey(t)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
	}))
	t.Cleanup(server.Close)
	t.Cleanup(server.CloseClientConnections)
	t.Cleanup(func() { close(release) })
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result, err := (&checker{tokenURL: server.URL + "/token", dnsURL: server.URL}).Check(ctx, validConfig(key))
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result, key)
}

func validConfig(serviceAccountKey string) map[string]any {
	return map[string]any{"projectId": projectId, "serviceAccountKey": serviceAccountKey}
}

func mustKey(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	pemText := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	raw, err := json.Marshal(map[string]string{
		"type":         "service_account",
		"client_email": clientEmail,
		"private_key":  pemText,
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
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

func assertNoSecrets(t *testing.T, got *core.AccessCheckResult, serviceAccountKey string) {
	t.Helper()
	joined := strings.Join(got.Diagnostics, "\n")
	for _, secret := range []string{accessToken, "PRIVATE KEY", clientEmail} {
		if strings.Contains(joined, secret) {
			t.Fatal("diagnostics leaked secret material")
		}
	}
	if strings.Contains(serviceAccountKey, "PRIVATE KEY") && strings.Contains(joined, serviceAccountKey) {
		t.Fatal("diagnostics leaked the service account key")
	}
}
