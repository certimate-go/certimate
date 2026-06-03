package apig

import (
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/byteplusquery"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/client"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/client/metadata"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/corehandlers"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/request"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/signer/byteplussign"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/service/apig"
)

type APIG struct {
	*client.Client
}

const (
	ServiceName = apig.ServiceName
	EndpointsID = apig.EndpointsID
	ServiceID   = apig.ServiceID
)

func New(p client.ConfigProvider, cfgs ...*byteplus.Config) *APIG {
	c := p.ClientConfig(EndpointsID, cfgs...)
	return newClient(*c.Config, c.Handlers, c.Endpoint, c.SigningRegion, c.SigningName)
}

func newClient(cfg byteplus.Config, handlers request.Handlers, endpoint, signingRegion, signingName string) *APIG {
	svc := &APIG{
		Client: client.New(
			cfg,
			metadata.ClientInfo{
				ServiceName:   ServiceName,
				ServiceID:     ServiceID,
				SigningName:   signingName,
				SigningRegion: signingRegion,
				Endpoint:      endpoint,
				APIVersion:    "2021-03-03",
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

func (c *APIG) newRequest(op *request.Operation, params, data interface{}) *request.Request {
	req := c.NewRequest(op, params, data)

	return req
}
