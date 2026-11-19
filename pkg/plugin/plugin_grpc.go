package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"

	githubplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	protolib "google.golang.org/protobuf/proto"

	"github.com/certimate-go/certimate/pkg/plugin/proto"
)

type frameSendFunc func(*proto.LogEntry) error

func (f frameSendFunc) Send(e *proto.LogEntry) error { return f(e) }

type GRPCPlugin struct {
	githubplugin.NetRPCUnsupportedPlugin

	Impl Plugin
}

func (p *GRPCPlugin) GRPCServer(_ *githubplugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterPluginServiceServer(s, &pluginServiceServer{impl: p.Impl, executions: newExecutionRegistry()})
	return nil
}

func (p *GRPCPlugin) GRPCClient(_ context.Context, _ *githubplugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &pluginServiceClient{raw: proto.NewPluginServiceClient(c)}, nil
}

type pluginExecution struct {
	postMu   sync.Mutex
	request  *proto.ProcessRequest
	done     chan struct{}
	postDone bool
	postErr  error
}

type executionRegistry struct {
	mu         sync.Mutex
	executions map[string]*pluginExecution
}

func newExecutionRegistry() *executionRegistry {
	return &executionRegistry{executions: make(map[string]*pluginExecution)}
}

func (r *executionRegistry) begin(id string, req *proto.ProcessRequest) *pluginExecution {
	r.mu.Lock()
	defer r.mu.Unlock()
	ex := &pluginExecution{request: protolib.Clone(req).(*proto.ProcessRequest), done: make(chan struct{})}
	r.executions[id] = ex
	return ex
}

func (r *executionRegistry) lookup(id string) *pluginExecution {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.executions[id]
}

func encodeMetadata(m *Metadata, err error) (*proto.GetMetadataResponse, error) {
	if err != nil {
		return nil, err
	}
	return &proto.GetMetadataResponse{
		ProviderType:         m.ProviderType,
		AccessProviderType:   m.AccessProviderType,
		Capability:           m.Capability,
		ProtocolVersion:      m.ProtocolVersion,
		Category:             m.Category,
		DisplayNameKey:       m.DisplayNameKey,
		AccessDisplayNameKey: m.AccessDisplayNameKey,
	}, nil
}

func decodeMetadata(resp *proto.GetMetadataResponse) *Metadata {
	return &Metadata{
		ProviderType:         resp.ProviderType,
		AccessProviderType:   resp.AccessProviderType,
		Capability:           resp.Capability,
		ProtocolVersion:      resp.ProtocolVersion,
		Category:             resp.Category,
		DisplayNameKey:       resp.DisplayNameKey,
		AccessDisplayNameKey: resp.AccessDisplayNameKey,
		SupportsAccessCheck:  resp.SupportsAccessCheck,
		RuntimeConfigJSON:    resp.RuntimeConfigJson,
	}
}

func validateRuntimeConfig(raw []byte) error {
	if len(raw) == 0 {
		return nil
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' || !json.Valid(trimmed) {
		return errors.New("plugin: runtime config must be a JSON object")
	}
	return nil
}

func implementsAccessChecker(impl Plugin) bool {
	_, ok := impl.(AccessChecker)
	return ok
}

func encodeConfigSchema(cs *ConfigSchema, err error) (*proto.GetConfigSchemaResponse, error) {
	if err != nil {
		return nil, err
	}
	i18n := make(map[string]*proto.StringMap, len(cs.I18n))
	for locale, entries := range cs.I18n {
		i18n[locale] = &proto.StringMap{Entries: entries}
	}
	return &proto.GetConfigSchemaResponse{
		AccessSchemaJson:   cs.AccessSchemaJSON,
		ProviderSchemaJson: cs.ProviderSchemaJSON,
		I18NResources:      i18n,
	}, nil
}

func decodeConfigSchema(resp *proto.GetConfigSchemaResponse) *ConfigSchema {
	i18n := make(map[string]map[string]string, len(resp.I18NResources))
	for locale, sm := range resp.I18NResources {
		i18n[locale] = sm.GetEntries()
	}
	return &ConfigSchema{
		AccessSchemaJSON:   resp.AccessSchemaJson,
		ProviderSchemaJSON: resp.ProviderSchemaJson,
		I18n:               i18n,
	}
}

func encodeRequest(req *Request) *proto.ProcessRequest {
	out := &proto.ProcessRequest{
		ExecutionId:        req.ExecutionID,
		Capability:         req.Capability,
		LogLevel:           req.LogLevel,
		AccessConfigJson:   req.AccessConfigJSON,
		ExtendedConfigJson: req.ExtendedConfigJSON,
	}
	switch {
	case req.Deploy != nil:
		out.Payload = &proto.ProcessRequest_Deploy{Deploy: &proto.DeployPayload{
			CertificatePem: req.Deploy.CertificatePEM,
			PrivateKeyPem:  req.Deploy.PrivateKeyPEM,
		}}
	case req.Dns != nil:
		out.Payload = &proto.ProcessRequest_Dns{Dns: &proto.DnsPayload{
			Domain:  req.Dns.Domain,
			Token:   req.Dns.Token,
			KeyAuth: req.Dns.KeyAuth,
		}}
	case req.Notify != nil:
		out.Payload = &proto.ProcessRequest_Notify{Notify: &proto.NotifyPayload{
			Subject: req.Notify.Subject,
			Message: req.Notify.Message,
		}}
	case req.Purge != nil:
		out.Payload = &proto.ProcessRequest_Purge{Purge: &proto.PurgePayload{
			ExpirySeconds: req.Purge.ExpirySeconds,
		}}
	}
	return out
}

func decodeRequest(req *proto.ProcessRequest) *Request {
	out := &Request{
		ExecutionID:        req.ExecutionId,
		Capability:         req.Capability,
		LogLevel:           req.LogLevel,
		AccessConfigJSON:   req.AccessConfigJson,
		ExtendedConfigJSON: req.ExtendedConfigJson,
	}
	switch p := req.Payload.(type) {
	case *proto.ProcessRequest_Deploy:
		out.Deploy = &DeployPayload{CertificatePEM: p.Deploy.GetCertificatePem(), PrivateKeyPEM: p.Deploy.GetPrivateKeyPem()}
	case *proto.ProcessRequest_Dns:
		out.Dns = &DnsPayload{Domain: p.Dns.GetDomain(), Token: p.Dns.GetToken(), KeyAuth: p.Dns.GetKeyAuth()}
	case *proto.ProcessRequest_Notify:
		out.Notify = &NotifyPayload{Subject: p.Notify.GetSubject(), Message: p.Notify.GetMessage()}
	case *proto.ProcessRequest_Purge:
		out.Purge = &PurgePayload{ExpirySeconds: p.Purge.GetExpirySeconds()}
	}
	return out
}

func encodeResult(res *Result) *proto.ProcessResult {
	switch {
	case res.Deploy != nil:
		return &proto.ProcessResult{Result: &proto.ProcessResult_Deploy{Deploy: &proto.DeployResult{ExtendedDataJson: res.Deploy.ExtendedDataJSON}}}
	case res.Dns != nil:
		return &proto.ProcessResult{Result: &proto.ProcessResult_Dns{Dns: &proto.DnsResult{}}}
	case res.Notify != nil:
		return &proto.ProcessResult{Result: &proto.ProcessResult_Notify{Notify: &proto.NotifyResult{ExtendedDataJson: res.Notify.ExtendedDataJSON}}}
	case res.Purge != nil:
		return &proto.ProcessResult{Result: &proto.ProcessResult_Purge{Purge: &proto.PurgeResult{Count: int64(res.Purge.Count), ExtendedDataJson: res.Purge.ExtendedDataJSON}}}
	}
	return nil
}

func decodeResult(res *proto.ProcessResult, capability string) (*Result, error) {
	out := &Result{}
	switch r := res.GetResult().(type) {
	case *proto.ProcessResult_Deploy:
		out.Deploy = &DeployResult{ExtendedDataJSON: r.Deploy.GetExtendedDataJson()}
	case *proto.ProcessResult_Dns:
		out.Dns = &DnsResult{}
	case *proto.ProcessResult_Notify:
		out.Notify = &NotifyResult{ExtendedDataJSON: r.Notify.GetExtendedDataJson()}
	case *proto.ProcessResult_Purge:
		count := r.Purge.GetCount()
		if count < 0 {
			return nil, fmt.Errorf("illegal purge result count %d", count)
		}
		out.Purge = &PurgeResult{Count: int(count), ExtendedDataJSON: r.Purge.GetExtendedDataJson()}
	default:
		return nil, fmt.Errorf("process result carries no result")
	}
	if err := out.Validate(capability); err != nil {
		return nil, err
	}
	return out, nil
}

func encodeAccessCheckResult(res *AccessCheckResult) *proto.CheckAccessResponse {
	out := &proto.CheckAccessResponse{
		Status:      res.Status,
		ReasonCode:  res.ReasonCode,
		Diagnostics: res.Diagnostics,
		Items:       make([]*proto.CheckAccessItem, len(res.Items)),
	}
	for i, item := range res.Items {
		out.Items[i] = &proto.CheckAccessItem{Id: item.ID, Status: item.Status}
	}
	return out
}

func decodeAccessCheckResult(resp *proto.CheckAccessResponse) (*AccessCheckResult, error) {
	out := &AccessCheckResult{
		Status:      resp.Status,
		ReasonCode:  resp.ReasonCode,
		Diagnostics: resp.Diagnostics,
		Items:       make([]AccessCheckItem, len(resp.Items)),
	}
	for i, item := range resp.Items {
		out.Items[i] = AccessCheckItem{ID: item.GetId(), Status: item.GetStatus()}
	}
	if err := out.Validate(); err != nil {
		return nil, err
	}
	return out, nil
}

type pluginServiceServer struct {
	proto.UnimplementedPluginServiceServer
	impl       Plugin
	executions *executionRegistry
}

func (s *pluginServiceServer) GetMetadata(ctx context.Context, _ *proto.GetMetadataRequest) (*proto.GetMetadataResponse, error) {
	resp, err := encodeMetadata(s.impl.GetMetadata(ctx))
	if err != nil {
		return nil, err
	}
	resp.SupportsAccessCheck = implementsAccessChecker(s.impl)
	if provider, ok := s.impl.(RuntimeConfigProvider); ok {
		config, err := provider.RuntimeConfig(ctx)
		if err != nil {
			return nil, err
		}
		if err := validateRuntimeConfig(config); err != nil {
			return nil, err
		}
		resp.RuntimeConfigJson = config
	}
	return resp, nil
}

func (s *pluginServiceServer) GetConfigSchema(ctx context.Context, _ *proto.GetConfigSchemaRequest) (*proto.GetConfigSchemaResponse, error) {
	return encodeConfigSchema(s.impl.GetConfigSchema(ctx))
}

func (s *pluginServiceServer) Process(req *proto.ProcessRequest, stream proto.PluginService_ProcessServer) error {
	decoded := decodeRequest(req)
	if err := decoded.Validate(); err != nil {
		return fmt.Errorf("plugin: reject process request: %w", err)
	}
	if s.impl == nil {
		return errors.New("plugin: no implementation configured")
	}
	ex := s.executions.begin(decoded.ExecutionID, req)
	defer close(ex.done)

	ctx := stream.Context()
	logger := NewStreamLogger(frameSendFunc(func(e *proto.LogEntry) error {
		return stream.Send(&proto.ProcessFrame{Frame: &proto.ProcessFrame_Log{Log: e}})
	}), parseLogLevel(req.LogLevel))
	res, err := s.impl.Process(ctx, decoded, logger)
	if err != nil {
		return err
	}
	if verr := res.Validate(decoded.Capability); verr != nil {
		return fmt.Errorf("plugin: reject process result: %w", verr)
	}
	return stream.Send(&proto.ProcessFrame{
		Frame: &proto.ProcessFrame_Result{Result: encodeResult(res)},
	})
}

func (s *pluginServiceServer) PostProcess(req *proto.PostProcessRequest, stream proto.PluginService_PostProcessServer) error {
	ex := s.executions.lookup(req.GetExecutionId())
	if ex == nil {
		return fmt.Errorf("plugin: unknown execution id %q", req.GetExecutionId())
	}
	if !protolib.Equal(ex.request, req.GetRequest()) {
		return fmt.Errorf("plugin: post process request does not match original execution %q", req.GetExecutionId())
	}

	ex.postMu.Lock()
	defer ex.postMu.Unlock()
	if ex.postDone {
		if ex.postErr != nil {
			return ex.postErr
		}
		return stream.Send(&proto.PostProcessFrame{Frame: &proto.PostProcessFrame_Result{Result: &proto.PostProcessResult{}}})
	}

	ctx := stream.Context()
	select {
	case <-ex.done:
	case <-ctx.Done():
		return fmt.Errorf("plugin: execution %q still running when the finalization budget expired: %w", req.GetExecutionId(), ctx.Err())
	}

	logger := NewStreamLogger(frameSendFunc(func(e *proto.LogEntry) error {
		return stream.Send(&proto.PostProcessFrame{Frame: &proto.PostProcessFrame_Log{Log: e}})
	}), parseLogLevel(req.GetRequest().GetLogLevel()))
	err := s.impl.PostProcess(ctx, decodeRequest(req.GetRequest()), logger)
	ex.postDone = true
	ex.postErr = err
	if err != nil {
		return err
	}
	return stream.Send(&proto.PostProcessFrame{
		Frame: &proto.PostProcessFrame_Result{Result: &proto.PostProcessResult{}},
	})
}

func (s *pluginServiceServer) CheckAccess(ctx context.Context, req *proto.CheckAccessRequest) (*proto.CheckAccessResponse, error) {
	checker, ok := s.impl.(AccessChecker)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "plugin does not implement access connectivity checks")
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	res, err := checker.CheckAccess(ctx, &AccessCheckRequest{AccessConfigJSON: req.GetAccessConfigJson()}, logger)
	if err != nil {
		return nil, err
	}
	if verr := res.Validate(); verr != nil {
		return nil, fmt.Errorf("plugin: reject access check result: %w", verr)
	}
	return encodeAccessCheckResult(res), nil
}

type pluginServiceClient struct {
	raw proto.PluginServiceClient
}

func (c *pluginServiceClient) GetMetadata(ctx context.Context) (*Metadata, error) {
	resp, err := c.raw.GetMetadata(ctx, &proto.GetMetadataRequest{})
	if err != nil {
		return nil, err
	}
	if err := validateRuntimeConfig(resp.GetRuntimeConfigJson()); err != nil {
		return nil, err
	}
	return decodeMetadata(resp), nil
}

func (c *pluginServiceClient) GetConfigSchema(ctx context.Context) (*ConfigSchema, error) {
	resp, err := c.raw.GetConfigSchema(ctx, &proto.GetConfigSchemaRequest{})
	if err != nil {
		return nil, err
	}
	return decodeConfigSchema(resp), nil
}

func (c *pluginServiceClient) Process(ctx context.Context, req *Request, logger *slog.Logger) (*Result, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("plugin: reject process request: %w", err)
	}
	stream, err := c.raw.Process(ctx, encodeRequest(req))
	if err != nil {
		return nil, err
	}
	var result *Result
	for {
		frame, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if result == nil {
					return nil, &ErrPluginProtocol{Reason: "stream ended without a final result"}
				}
				return result, nil
			}
			if result != nil {
				return nil, &ErrPluginProtocol{Reason: fmt.Sprintf("stream failed after the final result: %v", err)}
			}
			return nil, err
		}
		if result != nil {
			return nil, &ErrPluginProtocol{Reason: "frame received after the final result"}
		}
		switch f := frame.Frame.(type) {
		case *proto.ProcessFrame_Log:
			forwardLog(ctx, logger, f.Log)
		case *proto.ProcessFrame_Result:
			parsed, err := decodeResult(f.Result, req.Capability)
			if err != nil {
				return nil, &ErrPluginProtocol{Reason: err.Error()}
			}
			result = parsed
		default:
			return nil, &ErrPluginProtocol{Reason: "unknown process frame"}
		}
	}
}

func (c *pluginServiceClient) PostProcess(ctx context.Context, req *Request, logger *slog.Logger) error {
	stream, err := c.raw.PostProcess(ctx, &proto.PostProcessRequest{ExecutionId: req.ExecutionID, Request: encodeRequest(req)})
	if err != nil {
		return err
	}
	completed := false
	for {
		frame, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if !completed {
					return &ErrPluginProtocol{Reason: "stream ended without a finalization result"}
				}
				return nil
			}
			if completed {
				return &ErrPluginProtocol{Reason: fmt.Sprintf("stream failed after the finalization result: %v", err)}
			}
			return err
		}
		if completed {
			return &ErrPluginProtocol{Reason: "frame received after the finalization result"}
		}
		switch f := frame.Frame.(type) {
		case *proto.PostProcessFrame_Log:
			forwardLog(ctx, logger, f.Log)
		case *proto.PostProcessFrame_Result:
			completed = true
		default:
			return &ErrPluginProtocol{Reason: "unknown finalization frame"}
		}
	}
}

func (c *pluginServiceClient) CheckAccess(ctx context.Context, req *AccessCheckRequest) (*AccessCheckResult, error) {
	resp, err := c.raw.CheckAccess(ctx, &proto.CheckAccessRequest{AccessConfigJson: req.AccessConfigJSON})
	if err != nil {
		return nil, err
	}
	return decodeAccessCheckResult(resp)
}

func forwardLog(ctx context.Context, logger *slog.Logger, e *proto.LogEntry) {
	if logger == nil || e == nil {
		return
	}
	attrs := make([]any, 0, len(e.Data))
	for k, v := range e.Data {
		attrs = append(attrs, slog.String(k, v))
	}
	logger.Log(ctx, slog.Level(e.Level), e.Message, attrs...)
}
