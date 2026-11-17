package azure

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

const (
	tenantId     = "11111111-1111-1111-1111-111111111111"
	clientId     = "22222222-2222-2222-2222-222222222222"
	clientSecret = "probe-azure-client-secret"
	accessToken  = "probe-azure-access-token"
)

func TestNew_Identity(t *testing.T) {
	checker := New()
	if checker.ProviderType() != string(domain.AccessProviderTypeAzure) {
		t.Fatalf("provider type = %q, want %s", checker.ProviderType(), domain.AccessProviderTypeAzure)
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
		{"clientId": clientId, "clientSecret": clientSecret},
		{"tenantId": " ", "clientId": clientId, "clientSecret": clientSecret},
		{"tenantId": tenantId, "clientId": clientId},
		{"tenantId": tenantId + "/evil", "clientId": clientId, "clientSecret": clientSecret},
		{"tenantId": tenantId, "clientId": clientId, "clientSecret": clientSecret, "cloudName": "moon"},
		{"tenantId": tenantId, "clientId": 1, "clientSecret": clientSecret},
	}
	checker := &checker{authority: "http://127.0.0.1:1", arm: "http://127.0.0.1:1"}
	for _, config := range cases {
		_, err := checker.Check(context.Background(), config)
		if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
			t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
		}
		if err != nil && (strings.Contains(err.Error(), clientSecret) || strings.Contains(err.Error(), accessToken)) {
			t.Fatal("config error leaked secret material")
		}
	}
}

func TestChecker_Check_Passes(t *testing.T) {
	var sawToken, sawARM bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/oauth2/v2.0/token"):
			sawToken = true
			if strings.Contains(string(body), clientSecret) == false {
				t.Error("token request did not carry the client secret")
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"` + accessToken + `","token_type":"Bearer"}`))
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/subscriptions"):
			sawARM = true
			if r.Header.Get("Authorization") != "Bearer "+accessToken {
				t.Error("arm request did not use the acquired token")
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"value":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{authority: server.URL, arm: server.URL}).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertNoSecrets(t, result)
	if !sawToken || !sawARM {
		t.Fatalf("token=%v arm=%v, want both probes", sawToken, sawARM)
	}
}

func TestChecker_Check_InvalidClientIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_client","error_description":"bad secret"}`))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{authority: server.URL, arm: server.URL}).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertNoSecrets(t, result)
}

func TestChecker_Check_ArmForbiddenStillPasses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"` + accessToken + `"}`))
			return
		}
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"code":"AuthorizationFailed"}}`))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{authority: server.URL, arm: server.URL}).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertNoSecrets(t, result)
}

func TestChecker_Check_RefusedConnection(t *testing.T) {
	result, err := (&checker{authority: "http://" + closedAddr(t), arm: "http://127.0.0.1:1"}).Check(context.Background(), validConfig())
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

	result, err := (&checker{authority: server.URL, arm: server.URL}).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonTlsFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_TimeoutIsInconclusive(t *testing.T) {
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
	}))
	t.Cleanup(server.Close)
	t.Cleanup(server.CloseClientConnections)
	t.Cleanup(func() { close(release) })
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result, err := (&checker{authority: server.URL, arm: server.URL}).Check(ctx, validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func validConfig() map[string]any {
	return map[string]any{
		"tenantId":     tenantId,
		"clientId":     clientId,
		"clientSecret": clientSecret,
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
	for _, secret := range []string{clientSecret, accessToken} {
		if strings.Contains(joined, secret) {
			t.Fatal("diagnostics leaked secret material")
		}
	}
}
