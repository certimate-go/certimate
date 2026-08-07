package plugin

import (
	"context"
	"log/slog"

	githubplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/certimate-go/certimate/pkg/plugin/proto"
)

const PluginName = "deployer"

type Metadata struct {
	ProviderType         string
	AccessProviderType   string
	ProtocolVersion      uint32
	DeployCategory       string
	DeployDisplayNameKey string
	AccessDisplayNameKey string
}

type ConfigSchema struct {
	AccessSchemaJSON []byte
	DeploySchemaJSON []byte
	I18n             map[string]map[string]string
}

type DeployRequest struct {
	LogLevel           string
	AccessConfigJSON   string
	ExtendedConfigJSON string
	CertificatePEM     string
	PrivateKeyPEM      string
}

type DeployResult struct {
	ExtendedDataJSON string
}

type Deployer interface {
	GetMetadata(ctx context.Context) (*Metadata, error)
	GetConfigSchema(ctx context.Context) (*ConfigSchema, error)
	Deploy(ctx context.Context, req *DeployRequest, logger *slog.Logger) (*DeployResult, error)
}

type PluginSet = githubplugin.PluginSet

func PluginSetForDeployer() PluginSet {
	return PluginSet{PluginName: &DeployerGRPCPlugin{}}
}

type DeployerGRPCPlugin struct {
	githubplugin.NetRPCUnsupportedPlugin

	Impl Deployer
}

func (p *DeployerGRPCPlugin) GRPCServer(_ *githubplugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterDeployerPluginServer(s, &deployerGRPCServer{impl: p.Impl})
	return nil
}

func (p *DeployerGRPCPlugin) GRPCClient(_ context.Context, _ *githubplugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &deployerGRPCClient{raw: proto.NewDeployerPluginClient(c)}, nil
}

type deployerGRPCServer struct {
	proto.UnimplementedDeployerPluginServer
	impl Deployer
}

func (s *deployerGRPCServer) GetMetadata(ctx context.Context, _ *proto.GetMetadataRequest) (*proto.GetMetadataResponse, error) {
	m, err := s.impl.GetMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return &proto.GetMetadataResponse{
		ProviderType:         m.ProviderType,
		AccessProviderType:   m.AccessProviderType,
		ProtocolVersion:      m.ProtocolVersion,
		DeployCategory:       m.DeployCategory,
		DeployDisplayNameKey: m.DeployDisplayNameKey,
		AccessDisplayNameKey: m.AccessDisplayNameKey,
	}, nil
}

func (s *deployerGRPCServer) GetConfigSchema(ctx context.Context, _ *proto.GetConfigSchemaRequest) (*proto.GetConfigSchemaResponse, error) {
	cs, err := s.impl.GetConfigSchema(ctx)
	if err != nil {
		return nil, err
	}
	i18n := make(map[string]*proto.StringMap, len(cs.I18n))
	for locale, entries := range cs.I18n {
		i18n[locale] = &proto.StringMap{Entries: entries}
	}
	return &proto.GetConfigSchemaResponse{
		AccessSchemaJson: cs.AccessSchemaJSON,
		DeploySchemaJson: cs.DeploySchemaJSON,
		I18NResources:    i18n,
	}, nil
}

func (s *deployerGRPCServer) Deploy(req *proto.DeployRequest, stream proto.DeployerPlugin_DeployServer) error {
	ctx := stream.Context()
	logger := NewStreamLogger(stream, parseLogLevel(req.LogLevel))
	res, err := s.impl.Deploy(ctx, &DeployRequest{
		LogLevel:           req.LogLevel,
		AccessConfigJSON:   req.AccessConfigJson,
		ExtendedConfigJSON: req.ExtendedConfigJson,
		CertificatePEM:     req.CertificatePem,
		PrivateKeyPEM:      req.PrivateKeyPem,
	}, logger)
	if err != nil {
		return err
	}
	return stream.Send(&proto.DeployFrame{
		Frame: &proto.DeployFrame_Result{
			Result: &proto.DeployResponse{ExtendedDataJson: res.ExtendedDataJSON},
		},
	})
}

type deployerGRPCClient struct {
	raw proto.DeployerPluginClient
}

func (c *deployerGRPCClient) GetMetadata(ctx context.Context) (*Metadata, error) {
	resp, err := c.raw.GetMetadata(ctx, &proto.GetMetadataRequest{})
	if err != nil {
		return nil, err
	}
	return &Metadata{
		ProviderType:         resp.ProviderType,
		AccessProviderType:   resp.AccessProviderType,
		ProtocolVersion:      resp.ProtocolVersion,
		DeployCategory:       resp.DeployCategory,
		DeployDisplayNameKey: resp.DeployDisplayNameKey,
		AccessDisplayNameKey: resp.AccessDisplayNameKey,
	}, nil
}

func (c *deployerGRPCClient) GetConfigSchema(ctx context.Context) (*ConfigSchema, error) {
	resp, err := c.raw.GetConfigSchema(ctx, &proto.GetConfigSchemaRequest{})
	if err != nil {
		return nil, err
	}
	i18n := make(map[string]map[string]string, len(resp.I18NResources))
	for locale, sm := range resp.I18NResources {
		i18n[locale] = sm.GetEntries()
	}
	return &ConfigSchema{
		AccessSchemaJSON: resp.AccessSchemaJson,
		DeploySchemaJSON: resp.DeploySchemaJson,
		I18n:             i18n,
	}, nil
}

func (c *deployerGRPCClient) Deploy(ctx context.Context, req *DeployRequest, logger *slog.Logger) (*DeployResult, error) {
	stream, err := c.raw.Deploy(ctx, &proto.DeployRequest{
		LogLevel:           req.LogLevel,
		AccessConfigJson:   req.AccessConfigJSON,
		ExtendedConfigJson: req.ExtendedConfigJSON,
		CertificatePem:     req.CertificatePEM,
		PrivateKeyPem:      req.PrivateKeyPEM,
	})
	if err != nil {
		return nil, err
	}
	for {
		frame, err := stream.Recv()
		if err != nil {
			return nil, err
		}
		switch f := frame.Frame.(type) {
		case *proto.DeployFrame_Log:
			forwardLog(ctx, logger, f.Log)
		case *proto.DeployFrame_Result:
			return &DeployResult{ExtendedDataJSON: f.Result.ExtendedDataJson}, nil
		}
	}
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
