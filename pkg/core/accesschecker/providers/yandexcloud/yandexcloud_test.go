package yandexcloud

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	dnsv1 "github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
	iamv1 "github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
)

const (
	probeFolderId       = "b1gprobe_folder"
	probeKeyId          = "probe-key-id"
	probeServiceAccount = "ajeprobe_service_account"
	probeIamToken       = "probe-iam-token-value"
)

var (
	keyFixtureOnce sync.Once
	keyFixturePem  string
	keyFixtureJson string
)

func rsaPem(t *testing.T) string {
	t.Helper()
	keyFixtureOnce.Do(func() {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return
		}
		der, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			return
		}
		keyFixturePem = string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
		keyFixtureJson = fmt.Sprintf(`{"id":%q,"serviceAccountId":%q,"privateKey":%q}`, probeKeyId, probeServiceAccount, keyFixturePem)
	})
	if keyFixturePem == "" {
		t.Fatal("failed to generate a private key fixture")
	}
	return keyFixturePem
}

func serviceAccountKey(t *testing.T) string {
	t.Helper()
	rsaPem(t)
	return keyFixtureJson
}

func TestNew_Identity(t *testing.T) {
	checker := New()
	if checker.ProviderType() != string(domain.AccessProviderTypeYandexCloud) {
		t.Fatalf("provider type = %q, want %s", checker.ProviderType(), domain.AccessProviderTypeYandexCloud)
	}
	scope := checker.Scope()
	if len(scope) != 2 || scope[0] != core.AccessCheckItemConnect || scope[1] != core.AccessCheckItemAuth {
		t.Fatalf("scope = %v, want connect and auth", scope)
	}
}

func TestChecker_Check_ConfigInvalid(t *testing.T) {
	_, ed25519Key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ed25519Der, err := x509.MarshalPKCS8PrivateKey(ed25519Key)
	if err != nil {
		t.Fatal(err)
	}
	ed25519Pem := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: ed25519Der}))

	cases := []struct {
		name   string
		config map[string]any
	}{
		{"nil config", nil},
		{"empty config", map[string]any{}},
		{"missing folderId", stripField(validConfig(t), "folderId")},
		{"blank folderId", withField(validConfig(t), "folderId", "  ")},
		{"missing serviceAccountKey", stripField(validConfig(t), "serviceAccountKey")},
		{"blank serviceAccountKey", withField(validConfig(t), "serviceAccountKey", "   ")},
		{"malformed key json", withField(validConfig(t), "serviceAccountKey", "not json")},
		{"key without id", withField(validConfig(t), "serviceAccountKey", fmt.Sprintf(`{"serviceAccountId":%q,"privateKey":%q}`, probeServiceAccount, rsaPem(t)))},
		{"key without service account subject", withField(validConfig(t), "serviceAccountKey", fmt.Sprintf(`{"id":%q,"privateKey":%q}`, probeKeyId, rsaPem(t)))},
		{"key with non rsa pem", withField(validConfig(t), "serviceAccountKey", fmt.Sprintf(`{"id":%q,"serviceAccountId":%q,"privateKey":%q}`, probeKeyId, probeServiceAccount, ed25519Pem))},
	}

	checker := &checker{resolverEndpoint: "127.0.0.1:1"}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := checker.Check(context.Background(), c.config)
			if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
				t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
			}
			if err != nil && strings.Contains(err.Error(), rsaPem(t)) {
				t.Fatal("config error leaked secret material")
			}
		})
	}
}

func TestChecker_Check_Passes(t *testing.T) {
	addr := newGrpcServer(t,
		func(ctx context.Context, req *iamv1.CreateIamTokenRequest) (*iamv1.CreateIamTokenResponse, error) {
			if req.GetJwt() == "" {
				t.Error("token exchange request missing the signed jwt")
			}
			return &iamv1.CreateIamTokenResponse{IamToken: probeIamToken, ExpiresAt: timestamppb.New(time.Now().Add(time.Hour))}, nil
		},
		func(ctx context.Context, req *dnsv1.ListDnsZonesRequest) (*dnsv1.ListDnsZonesResponse, error) {
			if req.FolderId != probeFolderId {
				t.Errorf("folderId = %q, want %q", req.FolderId, probeFolderId)
			}
			if req.PageSize != 1 {
				t.Errorf("pageSize = %d, want 1", req.PageSize)
			}
			md, _ := metadata.FromIncomingContext(ctx)
			auth := md.Get("authorization")
			if len(auth) != 1 || !strings.HasPrefix(auth[0], "Bearer "+probeIamToken) {
				t.Errorf("authorization metadata = %v, want bearer %q", auth, probeIamToken)
			}
			return &dnsv1.ListDnsZonesResponse{}, nil
		},
	)

	result, err := (&checker{resolverEndpoint: addr}).Check(context.Background(), validConfig(t))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertDiagnostics(t, result, addr, "HTTP 200", "credential accepted", "permissions on specific resources were not checked")
	assertNoSecrets(t, result)
}

func TestChecker_Check_InvalidKeyIsAuthRejected(t *testing.T) {
	addr := newGrpcServer(t,
		func(ctx context.Context, req *iamv1.CreateIamTokenRequest) (*iamv1.CreateIamTokenResponse, error) {
			return nil, status.Error(codes.Unauthenticated, "the jwt signature is invalid")
		},
		nil,
	)

	result, err := (&checker{resolverEndpoint: addr}).Check(context.Background(), validConfig(t))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertNoSecrets(t, result)
}

func TestChecker_Check_ForbiddenIsInconclusive(t *testing.T) {
	addr := newGrpcServer(t, nil, func(ctx context.Context, req *dnsv1.ListDnsZonesRequest) (*dnsv1.ListDnsZonesResponse, error) {
		return nil, status.Error(codes.PermissionDenied, "no permission on this folder")
	})

	result, err := (&checker{resolverEndpoint: addr}).Check(context.Background(), validConfig(t))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonProbeForbidden, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_ServerErrorIsRemoteError(t *testing.T) {
	addr := newGrpcServer(t, nil, func(ctx context.Context, req *dnsv1.ListDnsZonesRequest) (*dnsv1.ListDnsZonesResponse, error) {
		return nil, status.Error(codes.Internal, "upstream failure")
	})

	result, err := (&checker{resolverEndpoint: addr}).Check(context.Background(), validConfig(t))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonRemoteError, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

func TestChecker_Check_RefusedConnection(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	listener.Close()

	result, err := (&checker{resolverEndpoint: addr}).Check(context.Background(), validConfig(t))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonConnectFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "tcp or dns failure")
	assertNoSecrets(t, result)
}

func TestChecker_Check_TimeoutIsInconclusive(t *testing.T) {
	addr := newGrpcServer(t,
		func(ctx context.Context, req *iamv1.CreateIamTokenRequest) (*iamv1.CreateIamTokenResponse, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		},
		nil,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result, err := (&checker{resolverEndpoint: addr}).Check(ctx, validConfig(t))
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result)
}

type fakeIamTokenServer struct {
	iamv1.UnimplementedIamTokenServiceServer
	create func(context.Context, *iamv1.CreateIamTokenRequest) (*iamv1.CreateIamTokenResponse, error)
}

func (f *fakeIamTokenServer) Create(ctx context.Context, req *iamv1.CreateIamTokenRequest) (*iamv1.CreateIamTokenResponse, error) {
	return f.create(ctx, req)
}

type fakeDnsZoneServer struct {
	dnsv1.UnimplementedDnsZoneServiceServer
	list func(context.Context, *dnsv1.ListDnsZonesRequest) (*dnsv1.ListDnsZonesResponse, error)
}

func (f *fakeDnsZoneServer) List(ctx context.Context, req *dnsv1.ListDnsZonesRequest) (*dnsv1.ListDnsZonesResponse, error) {
	return f.list(ctx, req)
}

func newGrpcServer(t *testing.T, create func(context.Context, *iamv1.CreateIamTokenRequest) (*iamv1.CreateIamTokenResponse, error), list func(context.Context, *dnsv1.ListDnsZonesRequest) (*dnsv1.ListDnsZonesResponse, error)) string {
	t.Helper()
	if create == nil {
		create = func(context.Context, *iamv1.CreateIamTokenRequest) (*iamv1.CreateIamTokenResponse, error) {
			return &iamv1.CreateIamTokenResponse{IamToken: probeIamToken, ExpiresAt: timestamppb.New(time.Now().Add(time.Hour))}, nil
		}
	}
	if list == nil {
		list = func(context.Context, *dnsv1.ListDnsZonesRequest) (*dnsv1.ListDnsZonesResponse, error) {
			return &dnsv1.ListDnsZonesResponse{}, nil
		}
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer()
	iamv1.RegisterIamTokenServiceServer(server, &fakeIamTokenServer{create: create})
	dnsv1.RegisterDnsZoneServiceServer(server, &fakeDnsZoneServer{list: list})
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)
	return listener.Addr().String()
}

func validConfig(t *testing.T) map[string]any {
	t.Helper()
	return map[string]any{
		"folderId":          probeFolderId,
		"serviceAccountKey": serviceAccountKey(t),
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
	if strings.Contains(joined, rsaPem(t)) {
		t.Fatal("diagnostics leaked secret material")
	}
}
