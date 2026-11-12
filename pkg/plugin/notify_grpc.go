package plugin

import (
	"context"

	githubplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/certimate-go/certimate/pkg/plugin/proto"
)

const NotifyPluginName = "notify"

type NotifyRequest struct {
	LogLevel           string
	AccessConfigJSON   string
	ExtendedConfigJSON string
	Subject            string
	Message            string
}

type NotifyResult struct {
	ExtendedDataJSON string
}

type Notifier interface {
	GetMetadata(ctx context.Context) (*Metadata, error)
	GetConfigSchema(ctx context.Context) (*ConfigSchema, error)
	Notify(ctx context.Context, req *NotifyRequest) (*NotifyResult, error)
}

func PluginSetForNotify() PluginSet {
	return PluginSet{NotifyPluginName: &NotifyGRPCPlugin{}}
}

type NotifyGRPCPlugin struct {
	githubplugin.NetRPCUnsupportedPlugin

	Impl Notifier
}

func (p *NotifyGRPCPlugin) GRPCServer(_ *githubplugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterNotifyPluginServer(s, &notifyGRPCServer{impl: p.Impl})
	return nil
}

func (p *NotifyGRPCPlugin) GRPCClient(_ context.Context, _ *githubplugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &notifyGRPCClient{raw: proto.NewNotifyPluginClient(c)}, nil
}

type notifyGRPCServer struct {
	proto.UnimplementedNotifyPluginServer
	impl Notifier
}

func (s *notifyGRPCServer) GetMetadata(ctx context.Context, _ *proto.GetMetadataRequest) (*proto.GetMetadataResponse, error) {
	return encodeMetadata(s.impl.GetMetadata(ctx))
}

func (s *notifyGRPCServer) GetConfigSchema(ctx context.Context, _ *proto.GetConfigSchemaRequest) (*proto.GetConfigSchemaResponse, error) {
	return encodeConfigSchema(s.impl.GetConfigSchema(ctx))
}

func (s *notifyGRPCServer) Notify(ctx context.Context, req *proto.NotifyRequest) (*proto.NotifyResponse, error) {
	res, err := s.impl.Notify(ctx, &NotifyRequest{
		LogLevel:           req.LogLevel,
		AccessConfigJSON:   req.AccessConfigJson,
		ExtendedConfigJSON: req.ExtendedConfigJson,
		Subject:            req.Subject,
		Message:            req.Message,
	})
	if err != nil {
		return nil, err
	}
	return &proto.NotifyResponse{ExtendedDataJson: res.ExtendedDataJSON}, nil
}

type notifyGRPCClient struct {
	raw proto.NotifyPluginClient
}

func (c *notifyGRPCClient) GetMetadata(ctx context.Context) (*Metadata, error) {
	resp, err := c.raw.GetMetadata(ctx, &proto.GetMetadataRequest{})
	if err != nil {
		return nil, err
	}
	return decodeMetadata(resp), nil
}

func (c *notifyGRPCClient) GetConfigSchema(ctx context.Context) (*ConfigSchema, error) {
	resp, err := c.raw.GetConfigSchema(ctx, &proto.GetConfigSchemaRequest{})
	if err != nil {
		return nil, err
	}
	return decodeConfigSchema(resp), nil
}

func (c *notifyGRPCClient) Notify(ctx context.Context, req *NotifyRequest) (*NotifyResult, error) {
	resp, err := c.raw.Notify(ctx, &proto.NotifyRequest{
		LogLevel:           req.LogLevel,
		AccessConfigJson:   req.AccessConfigJSON,
		ExtendedConfigJson: req.ExtendedConfigJSON,
		Subject:            req.Subject,
		Message:            req.Message,
	})
	if err != nil {
		return nil, err
	}
	return &NotifyResult{ExtendedDataJSON: resp.ExtendedDataJson}, nil
}
