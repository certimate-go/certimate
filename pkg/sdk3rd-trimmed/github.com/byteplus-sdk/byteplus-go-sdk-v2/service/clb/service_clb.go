package clb

import (
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/byteplusquery"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/client"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/client/metadata"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/corehandlers"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/request"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/signer/byteplussign"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/service/clb"
)

type CLB struct {
	*client.Client
}

const (
	ServiceName = clb.ServiceName
	EndpointsID = clb.EndpointsID
	ServiceID   = clb.ServiceID
)

func New(p client.ConfigProvider, cfgs ...*byteplus.Config) *CLB {
	c := p.ClientConfig(clb.EndpointsID, cfgs...)
	return newClient(*c.Config, c.Handlers, c.Endpoint, c.SigningRegion, c.SigningName)
}

func newClient(cfg byteplus.Config, handlers request.Handlers, endpoint, signingRegion, signingName string) *CLB {
	svc := &CLB{
		Client: client.New(
			cfg,
			metadata.ClientInfo{
				ServiceName:   clb.ServiceName,
				ServiceID:     clb.ServiceID,
				SigningName:   signingName,
				SigningRegion: signingRegion,
				Endpoint:      endpoint,
				APIVersion:    "2020-04-01",
			},
			handlers,
		),
	}

	svc.Handlers.Build.PushBackNamed(corehandlers.SDKVersionUserAgentHandler)
	svc.Handlers.Build.PushBackNamed(corehandlers.AddHostExecEnvUserAgentHandler)
	svc.Handlers.Sign.PushBackNamed(byteplussign.SignRequestHandler)
	svc.Handlers.Build.PushBackNamed(byteplusquery.BuildHandler)
	svc.Handlers.Unmarshal.PushBackNamed(byteplusquery.UnmarshalHandler)
	svc.Handlers.UnmarshalMeta.PushBackNamed(byteplusquery.UnmarshalMetaHandler)
	svc.Handlers.UnmarshalError.PushBackNamed(byteplusquery.UnmarshalErrorHandler)

	return svc
}

func (c *CLB) newRequest(op *request.Operation, params, data interface{}) *request.Request {
	req := c.NewRequest(op, params, data)

	return req
}
