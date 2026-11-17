package aws

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
	accessKeyId     = "AKIAprobeAwsAccessKey"
	secretAccessKey = "probe-aws-secret-value"
	imdsAccessKey   = "ASIAprobeImdsAccessKey"
	imdsSecret      = "probe-aws-imds-secret-value"
	imdsToken       = "probe-aws-imds-session-value"
)

const callerIdentityXML = `<GetCallerIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
  <GetCallerIdentityResult>
    <Arn>arn:aws:iam::123456789012:user/probe</Arn>
    <UserId>AIDACKCEVSQ6C2EXAMPLE</UserId>
    <Account>123456789012</Account>
  </GetCallerIdentityResult>
  <ResponseMetadata><RequestId>req-1</RequestId></ResponseMetadata>
</GetCallerIdentityResponse>`

const invalidClientXML = `<ErrorResponse>
  <Error>
    <Type>Sender</Type>
    <Code>InvalidClientTokenId</Code>
    <Message>the security token is invalid</Message>
  </Error>
  <RequestId>req-1</RequestId>
</ErrorResponse>`

func TestNew_Identity(t *testing.T) {
	checker := New()
	if checker.ProviderType() != string(domain.AccessProviderTypeAWS) {
		t.Fatalf("provider type = %q, want %s", checker.ProviderType(), domain.AccessProviderTypeAWS)
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
		{"authMethod": authMethodAccessKey},
		{"authMethod": authMethodAccessKey, "accessKeyId": " ", "secretAccessKey": secretAccessKey},
		{"authMethod": authMethodAccessKey, "accessKeyId": accessKeyId},
		{"authMethod": "sso", "accessKeyId": accessKeyId, "secretAccessKey": secretAccessKey},
		{"authMethod": authMethodAccessKey, "accessKeyId": 1, "secretAccessKey": secretAccessKey},
	}
	checker := &checker{stsEndpoint: "http://127.0.0.1:1"}
	for _, config := range cases {
		_, err := checker.Check(context.Background(), config)
		if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
			t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
		}
		if err != nil && (strings.Contains(err.Error(), accessKeyId) || strings.Contains(err.Error(), secretAccessKey)) {
			t.Fatal("config error leaked secret material")
		}
	}
}

func TestChecker_Check_AccessKeyPasses(t *testing.T) {
	var action string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		action = r.Form.Get("Action")
		w.Header().Set("Content-Type", "text/xml")
		_, _ = w.Write([]byte(callerIdentityXML))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{stsEndpoint: server.URL}).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertNoSecrets(t, result, accessKeyId, secretAccessKey)
	if action != "GetCallerIdentity" {
		t.Fatalf("action = %q, want GetCallerIdentity", action)
	}
}

func TestChecker_Check_InvalidKeyIsAuthRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(invalidClientXML))
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{stsEndpoint: server.URL}).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertNoSecrets(t, result, accessKeyId, secretAccessKey)
}

func TestChecker_Check_IMDSPasses(t *testing.T) {
	imdsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/latest/api/token":
			_, _ = w.Write([]byte("imds-probe-token"))
		case r.URL.Path == "/latest/meta-data/iam/security-credentials/" || r.URL.Path == "/latest/meta-data/iam/security-credentials":
			_, _ = w.Write([]byte("probe-role"))
		case strings.HasPrefix(r.URL.Path, "/latest/meta-data/iam/security-credentials/"):
			_, _ = w.Write([]byte(`{"Code":"Success","AccessKeyId":"` + imdsAccessKey + `","SecretAccessKey":"` + imdsSecret + `","Token":"` + imdsToken + `","Expiration":"2099-01-01T00:00:00Z"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(imdsServer.Close)

	stsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		_, _ = w.Write([]byte(callerIdentityXML))
	}))
	t.Cleanup(stsServer.Close)

	result, err := (&checker{stsEndpoint: stsServer.URL, imdsEndpoint: imdsServer.URL}).Check(context.Background(), map[string]any{"authMethod": authMethodIMDS})
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertNoSecrets(t, result, imdsAccessKey, imdsSecret, imdsToken)
}

func TestChecker_Check_RefusedConnection(t *testing.T) {
	result, err := (&checker{stsEndpoint: "http://" + closedAddr(t)}).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonConnectFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result, accessKeyId, secretAccessKey)
}

func TestChecker_Check_TlsFailure(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("the tls fixture must fail the handshake before any handler runs")
	}))
	t.Cleanup(server.Close)

	result, err := (&checker{stsEndpoint: server.URL}).Check(context.Background(), validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonTlsFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result, accessKeyId, secretAccessKey)
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

	result, err := (&checker{stsEndpoint: server.URL}).Check(ctx, validConfig())
	if err != nil {
		t.Fatal(err)
	}
	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result, accessKeyId, secretAccessKey)
}

func validConfig() map[string]any {
	return map[string]any{
		"authMethod":      authMethodAccessKey,
		"accessKeyId":     accessKeyId,
		"secretAccessKey": secretAccessKey,
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

func assertNoSecrets(t *testing.T, got *core.AccessCheckResult, secrets ...string) {
	t.Helper()
	joined := strings.Join(got.Diagnostics, "\n")
	for _, secret := range secrets {
		if secret != "" && strings.Contains(joined, secret) {
			t.Fatal("diagnostics leaked secret material")
		}
	}
}
