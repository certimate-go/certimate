package ovhcloud

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
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
	applicationKey    = "ovhprobeApplicationKey"
	applicationSecret = "probe-ovh-app-secret-value"
	consumerKey       = "ovhprobeConsumerKey"
	oauthClientId     = "ovhprobeClientId"
	oauthClientSecret = "probe-ovh-client-secret-value"
)

func TestNew_Identity(t *testing.T) {
	checker := New()
	if checker.ProviderType() != string(domain.AccessProviderTypeOVHcloud) {
		t.Fatalf("provider type = %q, want %s", checker.ProviderType(), domain.AccessProviderTypeOVHcloud)
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
		{"missing endpoint", applicationConfig("")},
		{"blank endpoint", applicationConfig("   ")},
		{"missing auth method", stripField(applicationConfig("http://127.0.0.1:1"), "authMethod")},
		{"unsupported auth method", withField(applicationConfig("http://127.0.0.1:1"), "authMethod", "apitoken")},
		{"application missing applicationKey", stripField(applicationConfig("http://127.0.0.1:1"), "applicationKey")},
		{"application missing applicationSecret", stripField(applicationConfig("http://127.0.0.1:1"), "applicationSecret")},
		{"application missing consumerKey", stripField(applicationConfig("http://127.0.0.1:1"), "consumerKey")},
		{"oauth2 missing clientId", stripField(oauth2Config("ovh-eu"), "clientId")},
		{"oauth2 missing clientSecret", stripField(oauth2Config("ovh-eu"), "clientSecret")},
		{"endpoint with trailing slash", applicationConfig("http://127.0.0.1:1/")},
		{"oauth2 with url endpoint", oauth2Config("http://127.0.0.1:1")},
	}

	probe := &checker{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := probe.Check(context.Background(), c.config)
			if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
				t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
			}
			if err != nil && (strings.Contains(err.Error(), applicationSecret) || strings.Contains(err.Error(), consumerKey) || strings.Contains(err.Error(), oauthClientSecret)) {
				t.Fatal("config error leaked secret material")
			}
		})
	}
}

func TestChecker_Check_Passes(t *testing.T) {
	server := newOvhServer(t, func(r *http.Request) (int, string) {
		if r.URL.Path == "/auth/time" {
			return http.StatusOK, fmt.Sprintf("%d", time.Now().Unix())
		}
		return http.StatusOK, `{"credentialId":"probe-1","status":"validated"}`
	})

	result, err := (&checker{}).Check(context.Background(), applicationConfig(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertDiagnostics(t, result, hostOf(server.URL), "HTTP 200", "credential accepted", "permissions on specific resources were not checked")
	assertNoSecrets(t, result)
}

func TestChecker_Check_InvalidCredentialIsAuthRejected(t *testing.T) {
	server := newOvhServer(t, func(r *http.Request) (int, string) {
		if r.URL.Path == "/auth/currentCredential" {
			return http.StatusUnauthorized, `{"class":"InvalidCredential","message":"this credential is not valid"}`
		}
		return http.StatusOK, fmt.Sprintf("%d", time.Now().Unix())
	})

	result, err := (&checker{}).Check(context.Background(), applicationConfig(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, "InvalidCredential")
	assertNoSecrets(t, result)
}

func TestChecker_Check_ForbiddenIsInconclusive(t *testing.T) {
	server := newOvhServer(t, func(r *http.Request) (int, string) {
		if r.URL.Path == "/auth/currentCredential" {
			return http.StatusForbidden, `{"class":"Forbidden","message":"not allowed"}`
		}
		return http.StatusOK, fmt.Sprintf("%d", time.Now().Unix())
	})

	result, err := (&checker{}).Check(context.Background(), applicationConfig(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonProbeForbidden, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_ServerErrorIsRemoteError(t *testing.T) {
	server := newOvhServer(t, func(r *http.Request) (int, string) {
		return http.StatusBadGateway, `{"class":"HttpError","message":"upstream failure"}`
	})

	result, err := (&checker{}).Check(context.Background(), applicationConfig(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_TimeServiceBrokenIsRemoteError(t *testing.T) {
	server := newOvhServer(t, func(r *http.Request) (int, string) {
		if r.URL.Path == "/auth/time" {
			return http.StatusInternalServerError, `{"class":"TimeError","message":"clock unavailable"}`
		}
		return http.StatusOK, "{}"
	})

	result, err := (&checker{}).Check(context.Background(), applicationConfig(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_RefusedConnection(t *testing.T) {
	result, err := (&checker{}).Check(context.Background(), applicationConfig("http://"+closedAddr(t)))
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

	result, err := (&checker{}).Check(context.Background(), applicationConfig(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonTlsFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "tls handshake failed")
	assertNoSecrets(t, result)
}

func TestChecker_Check_TimeoutIsInconclusive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/time" {
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, fmt.Sprintf("%d", time.Now().Unix()))
			return
		}
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result, err := (&checker{}).Check(ctx, applicationConfig(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func newOvhServer(t *testing.T, respond func(r *http.Request) (int, string)) *httptest.Server {
	t.Helper()
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/time" {
			status, body := respond(r)
			w.WriteHeader(status)
			_, _ = fmt.Fprint(w, body)
			return
		}

		if r.Header.Get("X-Ovh-Application") != applicationKey {
			t.Errorf("X-Ovh-Application = %q, want %q", r.Header.Get("X-Ovh-Application"), applicationKey)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Header.Get("X-Ovh-Consumer") != consumerKey {
			t.Errorf("X-Ovh-Consumer = %q, want %q", r.Header.Get("X-Ovh-Consumer"), consumerKey)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		target := "http://" + r.Host + r.URL.Path
		timestamp := r.Header.Get("X-Ovh-Timestamp")
		payload := fmt.Sprintf("%s+%s+%s+%s+%s+%s", applicationSecret, consumerKey, r.Method, target, "", timestamp)
		digest := sha1.Sum([]byte(payload))
		expected := "$1$" + hex.EncodeToString(digest[:])
		if signature := r.Header.Get("X-Ovh-Signature"); signature != expected {
			t.Errorf("X-Ovh-Signature = %q, want %q", signature, expected)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		status, body := respond(r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = fmt.Fprint(w, body)
	}))
	t.Cleanup(server.Close)
	return server
}

func applicationConfig(endpoint string) map[string]any {
	return map[string]any{
		"endpoint":          endpoint,
		"authMethod":        "application",
		"applicationKey":    applicationKey,
		"applicationSecret": applicationSecret,
		"consumerKey":       consumerKey,
	}
}

func oauth2Config(endpoint string) map[string]any {
	return map[string]any{
		"endpoint":     endpoint,
		"authMethod":   "oauth2",
		"clientId":     oauthClientId,
		"clientSecret": oauthClientSecret,
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
	for _, secret := range []string{applicationSecret, consumerKey, oauthClientSecret} {
		if strings.Contains(joined, secret) {
			t.Fatal("diagnostics leaked secret material")
		}
	}
}
