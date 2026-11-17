package plugin

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/certimate-go/certimate/pkg/plugin/proto"
)

type recordingPlugin struct {
	metadata     *Metadata
	configSchema *ConfigSchema

	processImpl func(ctx context.Context, req *Request, logger *slog.Logger) (*Result, error)
	postImpl    func(ctx context.Context, req *Request, logger *slog.Logger) error

	processCalls int
	postCalls    int
	processReq   *Request
	postReq      *Request
}

func (f *recordingPlugin) GetMetadata(context.Context) (*Metadata, error) {
	return f.metadata, nil
}

func (f *recordingPlugin) GetConfigSchema(context.Context) (*ConfigSchema, error) {
	return f.configSchema, nil
}

func (f *recordingPlugin) Process(ctx context.Context, req *Request, logger *slog.Logger) (*Result, error) {
	f.processCalls++
	f.processReq = req
	if f.processImpl != nil {
		return f.processImpl(ctx, req, logger)
	}
	return &Result{}, errors.New("process not configured")
}

func (f *recordingPlugin) PostProcess(ctx context.Context, req *Request, logger *slog.Logger) error {
	f.postCalls++
	f.postReq = req
	if f.postImpl != nil {
		return f.postImpl(ctx, req, logger)
	}
	return nil
}

func newRoundTrip(t *testing.T, impl Plugin) (*pluginServiceClient, func()) {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	proto.RegisterPluginServiceServer(srv, &pluginServiceServer{impl: impl, executions: newExecutionRegistry()})
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
	p := &GRPCPlugin{}
	raw, err := p.GRPCClient(ctx, nil, conn)
	if err != nil {
		t.Fatalf("grpc client: %v", err)
	}
	cleanup := func() {
		_ = conn.Close()
		srv.Stop()
	}
	return raw.(*pluginServiceClient), cleanup
}

func testMetadata(capability string) *Metadata {
	return &Metadata{
		ProviderType:         "fake",
		AccessProviderType:   "fake",
		Capability:           capability,
		ProtocolVersion:      ProtocolVersion,
		Category:             "other",
		DisplayNameKey:       "plugin.fake.name",
		AccessDisplayNameKey: "plugin.fake.name",
	}
}

func deployRequest() *Request {
	return &Request{
		Capability:         CapabilityDeploy,
		AccessConfigJSON:   `{"apiKey":"secret-value"}`,
		ExtendedConfigJSON: `{"endpoint":"https://example.com"}`,
		Deploy:             &DeployPayload{CertificatePEM: "-----BEGIN CERT-----", PrivateKeyPEM: "-----BEGIN KEY-----"},
	}
}

func dnsRequest(id string) *Request {
	return &Request{
		ExecutionID:        id,
		Capability:         CapabilityDns,
		AccessConfigJSON:   `{"apiKey":"secret-value"}`,
		ExtendedConfigJSON: `{}`,
		Dns:                &DnsPayload{Domain: "example.com", Token: "token-a", KeyAuth: "keyauth-a"},
	}
}

func notifyRequest() *Request {
	return &Request{
		Capability: CapabilityNotify,
		Notify:     &NotifyPayload{Subject: "hello", Message: "world"},
	}
}

func purgeRequest() *Request {
	return &Request{
		Capability: CapabilityPurge,
		Purge:      &PurgePayload{ExpirySeconds: 604800},
	}
}

func TestPluginGRPC_MetadataSchemaRoundTrip(t *testing.T) {
	impl := &recordingPlugin{
		metadata:     testMetadata(CapabilityDeploy),
		configSchema: &ConfigSchema{ProviderSchemaJSON: []byte(`{"x":1}`), I18n: map[string]map[string]string{"en": {"k": "v"}}},
	}
	client, cleanup := newRoundTrip(t, impl)
	defer cleanup()

	meta, err := client.GetMetadata(context.Background())
	if err != nil {
		t.Fatalf("GetMetadata: %v", err)
	}
	if meta.ProviderType != "fake" || meta.Capability != CapabilityDeploy || meta.Category != "other" || meta.DisplayNameKey != "plugin.fake.name" {
		t.Fatalf("metadata round trip mismatch: %+v", meta)
	}
	schema, err := client.GetConfigSchema(context.Background())
	if err != nil {
		t.Fatalf("GetConfigSchema: %v", err)
	}
	if string(schema.ProviderSchemaJSON) != `{"x":1}` || schema.I18n["en"]["k"] != "v" {
		t.Fatalf("schema round trip mismatch: %+v", schema)
	}
}

func TestPluginGRPC_ProcessAllCapabilities(t *testing.T) {
	cases := []struct {
		name      string
		capability string
		req       *Request
		verify    func(t *testing.T, res *Result)
	}{
		{
			name:       "deploy",
			capability: CapabilityDeploy,
			req:        deployRequest(),
			verify: func(t *testing.T, res *Result) {
				if res.Deploy == nil || res.Deploy.ExtendedDataJSON != `{"deployed":true}` {
					t.Fatalf("deploy result mismatch: %+v", res)
				}
			},
		},
		{
			name:       "dns",
			capability: CapabilityDns,
			req:        dnsRequest(""),
			verify: func(t *testing.T, res *Result) {
				if res.Dns == nil {
					t.Fatalf("dns result mismatch: %+v", res)
				}
			},
		},
		{
			name:       "notify",
			capability: CapabilityNotify,
			req:        notifyRequest(),
			verify: func(t *testing.T, res *Result) {
				if res.Notify == nil || res.Notify.ExtendedDataJSON != `{"sent":true}` {
					t.Fatalf("notify result mismatch: %+v", res)
				}
			},
		},
		{
			name:       "purge",
			capability: CapabilityPurge,
			req:        purgeRequest(),
			verify: func(t *testing.T, res *Result) {
				if res.Purge == nil || res.Purge.Count != 3 {
					t.Fatalf("purge result mismatch: %+v", res)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			impl := &recordingPlugin{metadata: testMetadata(tc.capability)}
			impl.processImpl = func(_ context.Context, req *Request, logger *slog.Logger) (*Result, error) {
				if logger != nil {
					logger.Info("processing", slog.String("capability", req.Capability))
				}
				switch req.Capability {
				case CapabilityDeploy:
					return &Result{Deploy: &DeployResult{ExtendedDataJSON: `{"deployed":true}`}}, nil
				case CapabilityDns:
					return &Result{Dns: &DnsResult{}}, nil
				case CapabilityNotify:
					return &Result{Notify: &NotifyResult{ExtendedDataJSON: `{"sent":true}`}}, nil
				case CapabilityPurge:
					return &Result{Purge: &PurgeResult{Count: 3}}, nil
				}
				return nil, fmt.Errorf("unexpected capability %q", req.Capability)
			}
			client, cleanup := newRoundTrip(t, impl)
			defer cleanup()

			res, err := client.Process(context.Background(), tc.req, slog.Default())
			if err != nil {
				t.Fatalf("Process(%s): %v", tc.name, err)
			}
			tc.verify(t, res)
			if impl.processCalls != 1 {
				t.Fatalf("expected exactly one process call, got %d", impl.processCalls)
			}
		})
	}
}

func TestPluginGRPC_ProcessRejectsInvalidRequestsBeforeVendorCall(t *testing.T) {
	cases := []struct {
		name string
		req  *Request
	}{
		{name: "wrong capability", req: &Request{Capability: CapabilityDeploy, Notify: &NotifyPayload{Subject: "x"}}},
		{name: "unknown capability", req: &Request{Capability: "magic", Deploy: &DeployPayload{}}},
		{name: "empty payload", req: &Request{Capability: CapabilityDeploy}},
		{name: "two payloads", req: &Request{Capability: CapabilityDeploy, Deploy: &DeployPayload{}, Notify: &NotifyPayload{}}},
		{name: "negative purge expiry", req: &Request{Capability: CapabilityPurge, Purge: &PurgePayload{ExpirySeconds: -1}}},
		{name: "dns missing key auth", req: &Request{Capability: CapabilityDns, Dns: &DnsPayload{Domain: "example.com", Token: "t"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			impl := &recordingPlugin{metadata: testMetadata(tc.req.Capability)}
			client, cleanup := newRoundTrip(t, impl)
			defer cleanup()

			_, err := client.Process(context.Background(), tc.req, nil)
			if err == nil {
				t.Fatalf("expected rejection, got success")
			}
			if impl.processCalls != 0 {
				t.Fatalf("vendor implementation must not be invoked for invalid requests")
			}
		})
	}
}

func TestPluginGRPC_ProcessRejectsWrongResultType(t *testing.T) {
	impl := &recordingPlugin{metadata: testMetadata(CapabilityDeploy)}
	impl.processImpl = func(context.Context, *Request, *slog.Logger) (*Result, error) {
		return &Result{Notify: &NotifyResult{}}, nil
	}
	client, cleanup := newRoundTrip(t, impl)
	defer cleanup()

	_, err := client.Process(context.Background(), deployRequest(), nil)
	if err == nil {
		t.Fatalf("expected mismatched result to fail")
	}
}

func TestPluginGRPC_ProcessNilResultRejected(t *testing.T) {
	impl := &recordingPlugin{metadata: testMetadata(CapabilityDeploy)}
	impl.processImpl = func(context.Context, *Request, *slog.Logger) (*Result, error) {
		return nil, nil
	}
	client, cleanup := newRoundTrip(t, impl)
	defer cleanup()

	_, err := client.Process(context.Background(), deployRequest(), nil)
	if err == nil {
		t.Fatalf("expected nil result to fail")
	}
}

type rogueServer struct {
	proto.UnimplementedPluginServiceServer
	process func(req *proto.ProcessRequest, stream proto.PluginService_ProcessServer) error
	post    func(req *proto.PostProcessRequest, stream proto.PluginService_PostProcessServer) error
	check   func(ctx context.Context, req *proto.CheckAccessRequest) (*proto.CheckAccessResponse, error)
}

func (r *rogueServer) Process(req *proto.ProcessRequest, stream proto.PluginService_ProcessServer) error {
	return r.process(req, stream)
}

func (r *rogueServer) PostProcess(req *proto.PostProcessRequest, stream proto.PluginService_PostProcessServer) error {
	return r.post(req, stream)
}

func (r *rogueServer) CheckAccess(ctx context.Context, req *proto.CheckAccessRequest) (*proto.CheckAccessResponse, error) {
	return r.check(ctx, req)
}

func newRogueRoundTrip(t *testing.T, rogue proto.PluginServiceServer) (*pluginServiceClient, func()) {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	proto.RegisterPluginServiceServer(srv, rogue)
	go func() { _ = srv.Serve(lis) }()
	conn, err := grpc.DialContext(
		context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("dial bufnet: %v", err)
	}
	p := &GRPCPlugin{}
	raw, err := p.GRPCClient(context.Background(), nil, conn)
	if err != nil {
		t.Fatalf("grpc client: %v", err)
	}
	return raw.(*pluginServiceClient), func() { _ = conn.Close(); srv.Stop() }
}

func TestPluginGRPC_ClientRejectsStreamViolations(t *testing.T) {
	sendResult := func(stream proto.PluginService_ProcessServer, result *proto.ProcessResult) error {
		return stream.Send(&proto.ProcessFrame{Frame: &proto.ProcessFrame_Result{Result: result}})
	}

	t.Run("missing result", func(t *testing.T) {
		rogue := &rogueServer{process: func(_ *proto.ProcessRequest, stream proto.PluginService_ProcessServer) error {
			return stream.Send(&proto.ProcessFrame{Frame: &proto.ProcessFrame_Log{Log: &proto.LogEntry{Message: "hi"}}})
		}}
		client, cleanup := newRogueRoundTrip(t, rogue)
		defer cleanup()
		_, err := client.Process(context.Background(), deployRequest(), nil)
		if !IsPluginProtocol(err) {
			t.Fatalf("expected protocol error for stream without result, got %v", err)
		}
	})

	t.Run("duplicate result", func(t *testing.T) {
		rogue := &rogueServer{process: func(_ *proto.ProcessRequest, stream proto.PluginService_ProcessServer) error {
			if err := sendResult(stream, &proto.ProcessResult{Result: &proto.ProcessResult_Deploy{Deploy: &proto.DeployResult{}}}); err != nil {
				return err
			}
			return sendResult(stream, &proto.ProcessResult{Result: &proto.ProcessResult_Deploy{Deploy: &proto.DeployResult{}}})
		}}
		client, cleanup := newRogueRoundTrip(t, rogue)
		defer cleanup()
		_, err := client.Process(context.Background(), deployRequest(), nil)
		if !IsPluginProtocol(err) {
			t.Fatalf("expected protocol error for duplicate result, got %v", err)
		}
	})

	t.Run("error status after result", func(t *testing.T) {
		rogue := &rogueServer{process: func(_ *proto.ProcessRequest, stream proto.PluginService_ProcessServer) error {
			if err := sendResult(stream, &proto.ProcessResult{Result: &proto.ProcessResult_Deploy{Deploy: &proto.DeployResult{}}}); err != nil {
				return err
			}
			return errors.New("boom after result")
		}}
		client, cleanup := newRogueRoundTrip(t, rogue)
		defer cleanup()
		_, err := client.Process(context.Background(), deployRequest(), nil)
		if !IsPluginProtocol(err) {
			t.Fatalf("expected protocol error for error status after result, got %v", err)
		}
	})

	t.Run("result frame for wrong capability", func(t *testing.T) {
		rogue := &rogueServer{process: func(_ *proto.ProcessRequest, stream proto.PluginService_ProcessServer) error {
			return sendResult(stream, &proto.ProcessResult{Result: &proto.ProcessResult_Notify{Notify: &proto.NotifyResult{}}})
		}}
		client, cleanup := newRogueRoundTrip(t, rogue)
		defer cleanup()
		_, err := client.Process(context.Background(), deployRequest(), nil)
		if !IsPluginProtocol(err) {
			t.Fatalf("expected protocol error for mismatched result, got %v", err)
		}
	})

	t.Run("negative purge count", func(t *testing.T) {
		rogue := &rogueServer{process: func(_ *proto.ProcessRequest, stream proto.PluginService_ProcessServer) error {
			return sendResult(stream, &proto.ProcessResult{Result: &proto.ProcessResult_Purge{Purge: &proto.PurgeResult{Count: -2}}})
		}}
		client, cleanup := newRogueRoundTrip(t, rogue)
		defer cleanup()
		_, err := client.Process(context.Background(), purgeRequest(), nil)
		if !IsPluginProtocol(err) {
			t.Fatalf("expected protocol error for negative count, got %v", err)
		}
	})
}

func TestPluginGRPC_PostProcessEmptyCompletion(t *testing.T) {
	impl := &recordingPlugin{metadata: testMetadata(CapabilityNotify)}
	impl.processImpl = func(context.Context, *Request, *slog.Logger) (*Result, error) {
		return &Result{Notify: &NotifyResult{}}, nil
	}
	client, cleanup := newRoundTrip(t, impl)
	defer cleanup()

	req := notifyRequest()
	if _, err := client.Process(context.Background(), req, nil); err != nil {
		t.Fatalf("Process: %v", err)
	}
	if err := client.PostProcess(context.Background(), req, nil); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}
	if impl.postCalls != 1 || impl.postReq.ExecutionID != req.ExecutionID || impl.postReq.Notify.Subject != "hello" {
		t.Fatalf("post process invocation mismatch: calls=%d req=%+v", impl.postCalls, impl.postReq)
	}
}

func TestPluginGRPC_PostProcessUnknownExecutionRejected(t *testing.T) {
	impl := &recordingPlugin{metadata: testMetadata(CapabilityDns)}
	client, cleanup := newRoundTrip(t, impl)
	defer cleanup()

	err := client.PostProcess(context.Background(), dnsRequest("never-seen"), nil)
	if err == nil {
		t.Fatalf("expected unknown execution rejection")
	}
	if impl.postCalls != 0 {
		t.Fatalf("vendor finalization must not run for unknown execution")
	}
}

func TestPluginGRPC_PostProcessRequestMismatchRejected(t *testing.T) {
	impl := &recordingPlugin{metadata: testMetadata(CapabilityDns)}
	impl.processImpl = func(context.Context, *Request, *slog.Logger) (*Result, error) {
		return &Result{Dns: &DnsResult{}}, nil
	}
	client, cleanup := newRoundTrip(t, impl)
	defer cleanup()

	if _, err := client.Process(context.Background(), dnsRequest("exec-1"), nil); err != nil {
		t.Fatalf("Process: %v", err)
	}
	tampered := dnsRequest("exec-1")
	tampered.Dns.Domain = "other.example.com"
	if err := client.PostProcess(context.Background(), tampered, nil); err == nil {
		t.Fatalf("expected request mismatch rejection")
	}
	if impl.postCalls != 0 {
		t.Fatalf("vendor finalization must not run for mismatched request")
	}
}

func TestPluginGRPC_PostProcessIdempotentPerExecution(t *testing.T) {
	impl := &recordingPlugin{metadata: testMetadata(CapabilityDns)}
	impl.processImpl = func(context.Context, *Request, *slog.Logger) (*Result, error) {
		return &Result{Dns: &DnsResult{}}, nil
	}
	client, cleanup := newRoundTrip(t, impl)
	defer cleanup()

	req := dnsRequest("exec-1")
	if _, err := client.Process(context.Background(), req, nil); err != nil {
		t.Fatalf("Process: %v", err)
	}
	if err := client.PostProcess(context.Background(), req, nil); err != nil {
		t.Fatalf("first PostProcess: %v", err)
	}
	if err := client.PostProcess(context.Background(), req, nil); err != nil {
		t.Fatalf("repeat PostProcess: %v", err)
	}
	if impl.postCalls != 1 {
		t.Fatalf("repeated finalization must replay the terminal outcome, got %d vendor calls", impl.postCalls)
	}
}

func TestPluginGRPC_PostProcessWaitsForProcessHandlerExit(t *testing.T) {
	release := make(chan struct{})
	started := make(chan struct{})
	impl := &recordingPlugin{metadata: testMetadata(CapabilityDns)}
	impl.processImpl = func(ctx context.Context, _ *Request, _ *slog.Logger) (*Result, error) {
		close(started)
		select {
		case <-release:
		case <-ctx.Done():
		}
		return &Result{Dns: &DnsResult{}}, nil
	}
	var postEntered chan struct{}
	impl.postImpl = func(context.Context, *Request, *slog.Logger) error {
		close(postEntered)
		return nil
	}
	client, cleanup := newRoundTrip(t, impl)
	defer cleanup()

	req := dnsRequest("exec-1")
	processDone := make(chan error, 1)
	go func() { _, err := client.Process(context.Background(), req, nil); processDone <- err }()
	<-started

	postEntered = make(chan struct{})
	postDone := make(chan error, 1)
	go func() { postDone <- client.PostProcess(context.Background(), req, nil) }()

	select {
	case <-postEntered:
		t.Fatalf("finalization must not run while the process handler is still running")
	case <-time.After(150 * time.Millisecond):
	}
	close(release)
	if err := <-processDone; err != nil {
		t.Fatalf("Process: %v", err)
	}
	if err := <-postDone; err != nil {
		t.Fatalf("PostProcess: %v", err)
	}
	<-postEntered
}

func TestPluginGRPC_PostProcessMissingCompletionFails(t *testing.T) {
	rogue := &rogueServer{
		process: func(_ *proto.ProcessRequest, stream proto.PluginService_ProcessServer) error {
			return stream.Send(&proto.ProcessFrame{Frame: &proto.ProcessFrame_Result{Result: &proto.ProcessResult{Result: &proto.ProcessResult_Dns{Dns: &proto.DnsResult{}}}}})
		},
		post: func(_ *proto.PostProcessRequest, stream proto.PluginService_PostProcessServer) error {
			return stream.Send(&proto.PostProcessFrame{Frame: &proto.PostProcessFrame_Log{Log: &proto.LogEntry{Message: "no completion"}}})
		},
	}
	client, cleanup := newRogueRoundTrip(t, rogue)
	defer cleanup()

	req := dnsRequest("exec-1")
	if _, err := client.Process(context.Background(), req, nil); err != nil {
		t.Fatalf("Process: %v", err)
	}
	err := client.PostProcess(context.Background(), req, nil)
	if !IsPluginProtocol(err) {
		t.Fatalf("expected protocol error for missing completion, got %v", err)
	}
}

func TestPluginGRPC_RequestCloneIsImmutable(t *testing.T) {
	req := dnsRequest("exec-1")
	clone := req.Clone()
	clone.Dns.Domain = "changed.example.com"
	if req.Dns.Domain != "example.com" {
		t.Fatalf("clone must not alias payload pointers")
	}
	if req.Clone().Clone() == nil {
		t.Fatalf("clone of clone must work")
	}
}
