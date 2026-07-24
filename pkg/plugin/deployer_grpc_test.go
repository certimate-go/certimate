package plugin

import (
	"context"
	"errors"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/certimate-go/certimate/pkg/plugin/proto"
)

type fakeDeployer struct {
	metadata      *Metadata
	configSchema  *ConfigSchema
	deployRequest *DeployRequest
	deployResult  *DeployResult
	deployErr     error
	metadataCalls int
	schemaCalls   int
	deployCalls   int
}

func (f *fakeDeployer) GetMetadata(ctx context.Context) (*Metadata, error) {
	f.metadataCalls++
	return f.metadata, nil
}

func (f *fakeDeployer) GetConfigSchema(ctx context.Context) (*ConfigSchema, error) {
	f.schemaCalls++
	return f.configSchema, nil
}

func (f *fakeDeployer) Deploy(ctx context.Context, req *DeployRequest) (*DeployResult, error) {
	f.deployCalls++
	f.deployRequest = req
	if f.deployErr != nil {
		return nil, f.deployErr
	}
	return f.deployResult, nil
}

func newRoundTrip(t *testing.T, impl Deployer) (Deployer, func()) {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	proto.RegisterDeployerPluginServer(srv, &deployerGRPCServer{impl: impl})
	go func() { _ = srv.Serve(lis) }()

	ctx := context.Background()
	conn, err := grpc.DialContext(
		ctx, "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("dial bufnet: %v", err)
	}
	p := &DeployerGRPCPlugin{}
	raw, err := p.GRPCClient(ctx, nil, conn)
	if err != nil {
		t.Fatalf("grpc client: %v", err)
	}
	cleanup := func() {
		_ = conn.Close()
		srv.Stop()
	}
	return raw.(Deployer), cleanup
}

func TestDeployerGRPC_GetMetadataRoundTrip(t *testing.T) {
	impl := &fakeDeployer{metadata: &Metadata{
		ProviderType:         "webhook-deployer",
		AccessProviderType:   "webhook-deployer",
		ProtocolVersion:      ProtocolVersion,
		DeployCategory:       "other",
		DeployDisplayNameKey: "plugin.webhook-deployer.name",
		AccessDisplayNameKey: "plugin.webhook-deployer.name",
	}}
	client, cleanup := newRoundTrip(t, impl)
	defer cleanup()

	got, err := client.GetMetadata(context.Background())
	if err != nil {
		t.Fatalf("GetMetadata: %v", err)
	}
	if got.ProviderType != "webhook-deployer" || got.ProtocolVersion != ProtocolVersion {
		t.Fatalf("metadata mismatch: %+v", got)
	}
	if impl.metadataCalls != 1 {
		t.Fatalf("expected 1 metadata call, got %d", impl.metadataCalls)
	}
}

func TestDeployerGRPC_GetConfigSchemaRoundTrip(t *testing.T) {
	impl := &fakeDeployer{configSchema: &ConfigSchema{
		AccessSchemaJSON: []byte(`{"schemaVersion":"form/v1","provider":"webhook-deployer","category":"access"}`),
		DeploySchemaJSON: []byte(`{"schemaVersion":"form/v1","provider":"webhook-deployer","category":"deploy"}`),
		I18n: map[string]map[string]string{
			"zh": {"plugin.webhook-deployer.name": "Webhook 部署"},
			"en": {"plugin.webhook-deployer.name": "Webhook Deployer"},
		},
	}}
	client, cleanup := newRoundTrip(t, impl)
	defer cleanup()

	got, err := client.GetConfigSchema(context.Background())
	if err != nil {
		t.Fatalf("GetConfigSchema: %v", err)
	}
	if string(got.AccessSchemaJSON) == "" {
		t.Fatal("empty access schema")
	}
	if got.I18n["zh"]["plugin.webhook-deployer.name"] != "Webhook 部署" {
		t.Fatalf("i18n zh mismatch: %v", got.I18n)
	}
	if got.I18n["en"]["plugin.webhook-deployer.name"] != "Webhook Deployer" {
		t.Fatalf("i18n en mismatch: %v", got.I18n)
	}
}

func TestDeployerGRPC_DeployRoundTrip(t *testing.T) {
	impl := &fakeDeployer{deployResult: &DeployResult{ExtendedDataJSON: `{"foo":"bar"}`}}
	client, cleanup := newRoundTrip(t, impl)
	defer cleanup()

	req := &DeployRequest{
		LogLevel:           "DEBUG",
		AccessConfigJSON:   `{"url":"https://example.com","secret":"tok"}`,
		ExtendedConfigJSON: `{"method":"POST","path":"/cert"}`,
		CertificatePEM:     "CERT",
		PrivateKeyPEM:      "KEY",
	}
	res, err := client.Deploy(context.Background(), req)
	if err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	if res.ExtendedDataJSON != `{"foo":"bar"}` {
		t.Fatalf("deploy result mismatch: %v", res)
	}
	if impl.deployRequest.CertificatePEM != "CERT" || impl.deployRequest.PrivateKeyPEM != "KEY" {
		t.Fatalf("cert/key not forwarded: %+v", impl.deployRequest)
	}
	if impl.deployRequest.AccessConfigJSON == "" || impl.deployRequest.ExtendedConfigJSON == "" {
		t.Fatalf("config not forwarded: %+v", impl.deployRequest)
	}
	if impl.deployRequest.LogLevel != "DEBUG" {
		t.Fatalf("log level not forwarded: %q", impl.deployRequest.LogLevel)
	}
}

func TestDeployerGRPC_PluginErrorPreserved(t *testing.T) {
	impl := &fakeDeployer{deployErr: errors.New("boom from plugin")}
	client, cleanup := newRoundTrip(t, impl)
	defer cleanup()

	_, err := client.Deploy(context.Background(), &DeployRequest{})
	if err == nil {
		t.Fatal("expected plugin error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %T: %v", err, err)
	}
	if st.Message() != "boom from plugin" {
		t.Fatalf("expected message preserved across RPC, got %q", st.Message())
	}
}
