package plugin

import (
	"context"

	githubplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/certimate-go/certimate/pkg/plugin/proto"
)

const DnsPluginName = "dns"

type DnsRequest struct {
	LogLevel           string
	AccessConfigJSON   string
	ExtendedConfigJSON string
	Domain             string
	Token              string
	KeyAuth            string
}

type DnsProvider interface {
	GetMetadata(ctx context.Context) (*Metadata, error)
	GetConfigSchema(ctx context.Context) (*ConfigSchema, error)
	Present(ctx context.Context, req *DnsRequest) error
	CleanUp(ctx context.Context, req *DnsRequest) error
}

func PluginSetForDns() PluginSet {
	return PluginSet{DnsPluginName: &DnsGRPCPlugin{}}
}

type DnsGRPCPlugin struct {
	githubplugin.NetRPCUnsupportedPlugin

	Impl DnsProvider
}

func (p *DnsGRPCPlugin) GRPCServer(_ *githubplugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterDnsPluginServer(s, &dnsGRPCServer{impl: p.Impl})
	return nil
}

func (p *DnsGRPCPlugin) GRPCClient(_ context.Context, _ *githubplugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &dnsGRPCClient{raw: proto.NewDnsPluginClient(c)}, nil
}

type dnsGRPCServer struct {
	proto.UnimplementedDnsPluginServer
	impl DnsProvider
}

func (s *dnsGRPCServer) GetMetadata(ctx context.Context, _ *proto.GetMetadataRequest) (*proto.GetMetadataResponse, error) {
	return encodeMetadata(s.impl.GetMetadata(ctx))
}

func (s *dnsGRPCServer) GetConfigSchema(ctx context.Context, _ *proto.GetConfigSchemaRequest) (*proto.GetConfigSchemaResponse, error) {
	return encodeConfigSchema(s.impl.GetConfigSchema(ctx))
}

func (s *dnsGRPCServer) Present(ctx context.Context, req *proto.DnsRequest) (*proto.DnsResponse, error) {
	if err := s.impl.Present(ctx, decodeDnsRequest(req)); err != nil {
		return nil, err
	}
	return &proto.DnsResponse{}, nil
}

func (s *dnsGRPCServer) CleanUp(ctx context.Context, req *proto.DnsRequest) (*proto.DnsResponse, error) {
	if err := s.impl.CleanUp(ctx, decodeDnsRequest(req)); err != nil {
		return nil, err
	}
	return &proto.DnsResponse{}, nil
}

type dnsGRPCClient struct {
	raw proto.DnsPluginClient
}

func (c *dnsGRPCClient) GetMetadata(ctx context.Context) (*Metadata, error) {
	resp, err := c.raw.GetMetadata(ctx, &proto.GetMetadataRequest{})
	if err != nil {
		return nil, err
	}
	return decodeMetadata(resp), nil
}

func (c *dnsGRPCClient) GetConfigSchema(ctx context.Context) (*ConfigSchema, error) {
	resp, err := c.raw.GetConfigSchema(ctx, &proto.GetConfigSchemaRequest{})
	if err != nil {
		return nil, err
	}
	return decodeConfigSchema(resp), nil
}

func (c *dnsGRPCClient) Present(ctx context.Context, req *DnsRequest) error {
	_, err := c.raw.Present(ctx, encodeDnsRequest(req))
	return err
}

func (c *dnsGRPCClient) CleanUp(ctx context.Context, req *DnsRequest) error {
	_, err := c.raw.CleanUp(ctx, encodeDnsRequest(req))
	return err
}

func decodeDnsRequest(req *proto.DnsRequest) *DnsRequest {
	return &DnsRequest{
		LogLevel:           req.LogLevel,
		AccessConfigJSON:   req.AccessConfigJson,
		ExtendedConfigJSON: req.ExtendedConfigJson,
		Domain:             req.Domain,
		Token:              req.Token,
		KeyAuth:            req.KeyAuth,
	}
}

func encodeDnsRequest(req *DnsRequest) *proto.DnsRequest {
	return &proto.DnsRequest{
		LogLevel:           req.LogLevel,
		AccessConfigJson:   req.AccessConfigJSON,
		ExtendedConfigJson: req.ExtendedConfigJSON,
		Domain:             req.Domain,
		Token:              req.Token,
		KeyAuth:            req.KeyAuth,
	}
}
