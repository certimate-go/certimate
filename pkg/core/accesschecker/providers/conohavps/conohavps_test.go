package conohavps

import (
	"context"
	"encoding/json"
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
	valUserId    = "probe-conoha-user-id"
	valUserName  = "probe-conoha-user-name"
	valPassword  = "probe-conoha-password"
	valTenantId  = "probe-conoha-tenant-id"
	valTenantNm  = "probe-conoha-tenant-name"
	valIssuedTok = "probe-conoha-issued-token"
)

func TestNew_Identity(t *testing.T) {
	checker := New()
	if checker.ProviderType() != string(domain.AccessProviderTypeConoHaVPS) {
		t.Fatalf("provider type = %q, want %s", checker.ProviderType(), domain.AccessProviderTypeConoHaVPS)
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
		{"missing apiVersion", map[string]any{"apiUserName": valUserName, "apiPassword": valPassword, "tenantId": valTenantId}},
		{"unsupported apiVersion", map[string]any{"apiVersion": "v4", "apiUserName": valUserName, "apiPassword": valPassword, "tenantId": valTenantId}},
		{"v2 missing apiUserName", map[string]any{"apiVersion": "v2", "apiPassword": valPassword, "tenantId": valTenantId}},
		{"v2 missing tenantId", map[string]any{"apiVersion": "v2", "apiUserName": valUserName, "apiPassword": valPassword}},
		{"v3 missing user identifiers", map[string]any{"apiVersion": "v3", "apiPassword": valPassword, "tenantId": valTenantId}},
		{"v3 missing tenant identifiers", map[string]any{"apiVersion": "v3", "apiUserName": valUserName, "apiPassword": valPassword}},
		{"missing apiPassword", map[string]any{"apiVersion": "v2", "apiUserName": valUserName, "tenantId": valTenantId}},
		{"blank apiPassword", map[string]any{"apiVersion": "v2", "apiUserName": valUserName, "apiPassword": "   ", "tenantId": valTenantId}},
		{"malformed apiVersion", map[string]any{"apiVersion": 1, "apiUserName": valUserName, "apiPassword": valPassword, "tenantId": valTenantId}},
	}

	checker := &checker{v2IdentityURL: "http://127.0.0.1:1/v2.0/tokens", v3IdentityURL: "http://127.0.0.1:1/v3/auth/tokens"}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := checker.Check(context.Background(), c.config)
			if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
				t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
			}
			if err != nil && (strings.Contains(err.Error(), valPassword) || strings.Contains(err.Error(), valUserName)) {
				t.Fatal("config error leaked secret material")
			}
		})
	}
}

func TestChecker_Check_V2IssuesTokenAndPasses(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access":{"token":{"issued_at":"` + time.Now().UTC().Format(time.RFC3339) + `","expires":"2030-01-01T00:00:00Z","id":"` + valIssuedTok + `"}}}`))
	}))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL+"/v2.0/tokens", server.URL+"/v3/auth/tokens").Check(context.Background(), map[string]any{
		"apiVersion":  "v2",
		"apiUserName": valUserName,
		"apiPassword": valPassword,
		"tenantId":    valTenantId,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertDiagnostics(t, result, hostOf(server.URL), "HTTP 200", "token issued")
	assertNoSecrets(t, result)
	if gotPath != "/v2.0/tokens" {
		t.Fatalf("path = %q, want /v2.0/tokens", gotPath)
	}
	auth, _ := gotBody["auth"].(map[string]any)
	if auth == nil {
		t.Fatalf("body missing auth: %v", gotBody)
	}
	creds, _ := auth["passwordCredentials"].(map[string]any)
	if creds == nil || creds["username"] != valUserName || creds["password"] != valPassword {
		t.Fatalf("body missing password credentials: %v", gotBody)
	}
	if auth["tenantId"] != valTenantId {
		t.Fatalf("body missing tenantId: %v", gotBody)
	}
}

func TestChecker_Check_V3IssuesTokenAndPasses(t *testing.T) {
	var gotPath string
	var gotAuth map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		var payload struct {
			Auth map[string]any `json:"auth"`
		}
		json.Unmarshal(raw, &payload)
		gotAuth = payload.Auth
		w.Header().Set("X-Subject-Token", valIssuedTok)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"token":{"issued_at":"2026-01-01T00:00:00Z","expires":"2030-01-01T00:00:00Z"}}`))
	}))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL+"/v2.0/tokens", server.URL+"/v3/auth/tokens").Check(context.Background(), map[string]any{
		"apiVersion":  "v3",
		"apiUserId":   valUserId,
		"apiUserName": valUserName,
		"apiPassword": valPassword,
		"tenantId":    valTenantId,
		"tenantName":  valTenantNm,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertDiagnostics(t, result, hostOf(server.URL), "HTTP 201", "token issued")
	assertNoSecrets(t, result)
	if gotPath != "/v3/auth/tokens" {
		t.Fatalf("path = %q, want /v3/auth/tokens", gotPath)
	}
	identity, _ := gotAuth["identity"].(map[string]any)
	if identity == nil {
		t.Fatalf("body missing identity: %v", gotAuth)
	}
	password, _ := identity["password"].(map[string]any)
	user, _ := password["user"].(map[string]any)
	if user == nil || user["password"] != valPassword {
		t.Fatalf("body missing user password: %v", gotAuth)
	}
	if _, scoped := identity["scope"]; scoped {
		t.Fatalf("scope must be a sibling of identity per keystone v3: %v", gotAuth)
	}
	scope, _ := gotAuth["scope"].(map[string]any)
	project, _ := scope["project"].(map[string]any)
	if project == nil || project["id"] != valTenantId || project["name"] != valTenantNm {
		t.Fatalf("body missing scoped project: %v", gotAuth)
	}
}

func TestChecker_Check_NumericVersionAliasesRouteToV3(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"token":{"expires":"2030-01-01T00:00:00Z"}}`))
	}))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL+"/v2.0/tokens", server.URL+"/v3/auth/tokens").Check(context.Background(), map[string]any{
		"apiVersion":  "3",
		"apiUserName": valUserName,
		"apiPassword": valPassword,
		"tenantName":  valTenantNm,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	if gotPath != "/v3/auth/tokens" {
		t.Fatalf("path = %q, want /v3/auth/tokens", gotPath)
	}
}

func TestChecker_Check_V2UnauthorizedIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"The request you have made requires authentication.","code":401,"title":"Unauthorized"}}`))
	}))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL+"/v2.0/tokens", server.URL+"/v3/auth/tokens").Check(context.Background(), map[string]any{
		"apiVersion":  "v2",
		"apiUserName": valUserName,
		"apiPassword": valPassword,
		"tenantId":    valTenantId,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, "HTTP 401")
	assertNoSecrets(t, result)
}

func TestChecker_Check_V3UnauthorizedIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"The request you have made requires authentication.","code":401,"title":"Unauthorized"}}`))
	}))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL+"/v2.0/tokens", server.URL+"/v3/auth/tokens").Check(context.Background(), map[string]any{
		"apiVersion":  "v3",
		"apiUserId":   valUserId,
		"apiPassword": valPassword,
		"tenantId":    valTenantId,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertNoSecrets(t, result)
}

func TestChecker_Check_ServerErrorIsRemoteError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)

	result, err := checkerFor(t, server.URL+"/v2.0/tokens", server.URL+"/v3/auth/tokens").Check(context.Background(), map[string]any{
		"apiVersion":  "v2",
		"apiUserName": valUserName,
		"apiPassword": valPassword,
		"tenantId":    valTenantId,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "HTTP 502")
	assertNoSecrets(t, result)
}

func TestChecker_Check_RefusedConnection(t *testing.T) {
	addr := closedAddr(t)
	result, err := checkerFor(t, "http://"+addr+"/v2.0/tokens", "http://"+addr+"/v3/auth/tokens").Check(context.Background(), map[string]any{
		"apiVersion":  "v2",
		"apiUserName": valUserName,
		"apiPassword": valPassword,
		"tenantId":    valTenantId,
	})
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

	result, err := checkerFor(t, server.URL+"/v2.0/tokens", server.URL+"/v3/auth/tokens").Check(context.Background(), map[string]any{
		"apiVersion":  "v2",
		"apiUserName": valUserName,
		"apiPassword": valPassword,
		"tenantId":    valTenantId,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonTlsFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "tls handshake failed")
	assertNoSecrets(t, result)
}

func TestChecker_Check_TimeoutIsInconclusive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	}))
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result, err := checkerFor(t, server.URL+"/v2.0/tokens", server.URL+"/v3/auth/tokens").Check(ctx, map[string]any{
		"apiVersion":  "v2",
		"apiUserName": valUserName,
		"apiPassword": valPassword,
		"tenantId":    valTenantId,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func checkerFor(t *testing.T, v2URL, v3URL string) *checker {
	t.Helper()
	return &checker{v2IdentityURL: v2URL, v3IdentityURL: v3URL}
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
	for _, secret := range []string{valUserId, valUserName, valPassword, valTenantId, valTenantNm, valIssuedTok} {
		if strings.Contains(joined, secret) {
			t.Fatal("diagnostics leaked secret material")
		}
	}
}
