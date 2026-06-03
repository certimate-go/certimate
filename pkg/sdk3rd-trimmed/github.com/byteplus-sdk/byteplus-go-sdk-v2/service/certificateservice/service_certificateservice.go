package certificateservice

import (
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/byteplusquery"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/client"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/client/metadata"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/corehandlers"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/request"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/signer/byteplussign"
)

type CERTIFICATESERVICE struct {
	*client.Client
}

const (
	ServiceName = "certificate_service"
	EndpointsID = ServiceName
	ServiceID   = "certificate_service"
)

func New(p client.ConfigProvider, cfgs ...*byteplus.Config) *CERTIFICATESERVICE {
	c := p.ClientConfig(EndpointsID, cfgs...)
	return newClient(*c.Config, c.Handlers, c.Endpoint, c.SigningRegion, c.SigningName)
}

func newClient(cfg byteplus.Config, handlers request.Handlers, endpoint, signingRegion, signingName string) *CERTIFICATESERVICE {
	svc := &CERTIFICATESERVICE{
		Client: client.New(
			cfg,
			metadata.ClientInfo{
				ServiceName:   ServiceName,
				ServiceID:     ServiceID,
				SigningName:   signingName,
				SigningRegion: signingRegion,
				Endpoint:      endpoint,
				APIVersion:    "2021-06-01",
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

func (c *CERTIFICATESERVICE) newRequest(op *request.Operation, params, data interface{}) *request.Request {
	req := c.NewRequest(op, params, data)

	return req
}
