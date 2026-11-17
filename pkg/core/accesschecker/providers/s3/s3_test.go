package s3

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	miniogo "github.com/minio/minio-go/v7"

	"github.com/certimate-go/certimate/pkg/core"
)

func TestMain(m *testing.M) {
	prev := miniogo.MaxRetry
	miniogo.MaxRetry = 1
	code := m.Run()
	miniogo.MaxRetry = prev
	os.Exit(code)
}

const (
	accessKey = "AKIATESTKEY999"
	secretKey = "wJalrSECRETKEY999"

	listBucketsXML = `<?xml version="1.0" encoding="UTF-8"?>
<ListAllMyBucketsResult>
  <Buckets></Buckets>
</ListAllMyBucketsResult>`
)

func TestChecker_Check_ConfigInvalid(t *testing.T) {
	cases := []struct {
		name   string
		config map[string]any
	}{
		{"empty config", map[string]any{}},
		{"missing endpoint", map[string]any{"accessKey": accessKey, "secretKey": secretKey}},
		{"blank endpoint", map[string]any{"endpoint": "  ", "accessKey": accessKey, "secretKey": secretKey}},
		{"missing accessKey", map[string]any{"endpoint": "http://127.0.0.1:1", "secretKey": secretKey}},
		{"missing secretKey", map[string]any{"endpoint": "http://127.0.0.1:1", "accessKey": accessKey}},
		{"invalid signatureVersion", map[string]any{"endpoint": "http://127.0.0.1:1", "accessKey": accessKey, "secretKey": secretKey, "signatureVersion": "v1"}},
		{"malformed usePathStyle", map[string]any{"endpoint": "http://127.0.0.1:1", "accessKey": accessKey, "secretKey": secretKey, "usePathStyle": "yes"}},
	}

	checker := &checker{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := checker.Check(context.Background(), c.config)
			if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
				t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
			}
			if strings.Contains(err.Error(), secretKey) || strings.Contains(err.Error(), accessKey) {
				t.Fatalf("config error leaked credential material: %v", err)
			}
		})
	}
}

func TestChecker_Check_ListBucketsPasses(t *testing.T) {
	var auth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(listBucketsXML))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{}).Check(context.Background(), s3Config(server.URL, "v4", false))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertDiagnostics(t, result, hostOf(server.URL), "credential accepted")
	assertNoSecrets(t, result)
	if !strings.HasPrefix(auth, "AWS4-HMAC-SHA256") {
		t.Fatalf("expected a signature v4 authorization prefix")
	}
}

func TestChecker_Check_SignatureV2Passes(t *testing.T) {
	var auth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(listBucketsXML))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{}).Check(context.Background(), s3Config(server.URL, "V2", true))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertNoSecrets(t, result)
	if !strings.HasPrefix(auth, "AWS ") {
		t.Fatalf("expected a signature v2 authorization prefix")
	}
}

func TestChecker_Check_ForbiddenIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{}).Check(context.Background(), s3Config(server.URL, "", false))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, hostOf(server.URL), "HTTP 403")
	assertNoSecrets(t, result)
}

func TestChecker_Check_BadRequestIsRemoteError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{}).Check(context.Background(), s3Config(server.URL, "v4", false))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_TlsFailure(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("the tls fixture must fail the handshake before any handler runs")
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{}).Check(context.Background(), s3Config(server.URL, "v4", false))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonTlsFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, hostOf(server.URL), "tls handshake failed")
	assertNoSecrets(t, result)
}

func TestChecker_Check_InsecureTLSPasses(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(listBucketsXML))
	}))
	t.Cleanup(server.Close)

	config := s3Config(server.URL, "v4", false)
	config["allowInsecureConnections"] = true

	result, err := (&checker{}).Check(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertNoSecrets(t, result)
}

func TestChecker_Check_RefusedConnection(t *testing.T) {
	result, err := (&checker{}).Check(context.Background(), s3Config("http://"+closedAddr(t), "v4", false))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonConnectFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "tcp or dns failure")
	assertNoSecrets(t, result)
}

func TestChecker_Check_TimeoutIsInconclusive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result, err := (&checker{}).Check(ctx, s3Config(server.URL, "v4", false))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func s3Config(endpoint, signature string, pathStyle bool) map[string]any {
	config := map[string]any{
		"endpoint":     endpoint,
		"accessKey":    accessKey,
		"secretKey":    secretKey,
		"usePathStyle": pathStyle,
	}
	if signature != "" {
		config["signatureVersion"] = signature
	}
	return config
}

func hostOf(raw string) string {
	return endpointHost(raw)
}

func closedAddr(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()
	return addr
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
	for _, secret := range []string{accessKey, secretKey} {
		if strings.Contains(joined, secret) {
			t.Errorf("diagnostics leaked credential material: %v", got.Diagnostics)
		}
	}
}
