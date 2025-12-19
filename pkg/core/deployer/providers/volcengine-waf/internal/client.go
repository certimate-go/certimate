package internal

import (
	"github.com/volcengine/volcengine-go-sdk/service/waf"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/client"
	"github.com/volcengine/volcengine-go-sdk/volcengine/client/metadata"
	"github.com/volcengine/volcengine-go-sdk/volcengine/corehandlers"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
	"github.com/volcengine/volcengine-go-sdk/volcengine/signer/volc"
	"github.com/volcengine/volcengine-go-sdk/volcengine/volcenginequery"
)

// This is a partial copy of https://github.com/volcengine/volcengine-go-sdk/blob/master/service/waf/service_waf.go
// to lightweight the vendor packages in the built binary.
type WafClient struct {
	*client.Client
}

func NewWafClient(p client.ConfigProvider, cfgs ...*volcengine.Config) *WafClient {
	c := p.ClientConfig(waf.EndpointsID, cfgs...)
	return newDcdnClient(*c.Config, c.Handlers, c.Endpoint, c.SigningRegion, c.SigningName)
}

func newDcdnClient(cfg volcengine.Config, handlers request.Handlers, endpoint, signingRegion, signingName string) *WafClient {
	svc := &WafClient{
		Client: client.New(
			cfg,
			metadata.ClientInfo{
				ServiceName:   waf.ServiceName,
				ServiceID:     waf.ServiceID,
				SigningName:   signingName,
				SigningRegion: signingRegion,
				Endpoint:      endpoint,
				APIVersion:    "2023-12-25",
			},
			handlers,
		),
	}

	svc.Handlers.Build.PushBackNamed(corehandlers.SDKVersionUserAgentHandler)
	svc.Handlers.Build.PushBackNamed(corehandlers.AddHostExecEnvUserAgentHandler)
	svc.Handlers.Sign.PushBackNamed(volc.SignRequestHandler)
	svc.Handlers.Build.PushBackNamed(volcenginequery.BuildHandler)
	svc.Handlers.Unmarshal.PushBackNamed(volcenginequery.UnmarshalHandler)
	svc.Handlers.UnmarshalMeta.PushBackNamed(volcenginequery.UnmarshalMetaHandler)
	svc.Handlers.UnmarshalError.PushBackNamed(volcenginequery.UnmarshalErrorHandler)

	return svc
}

func (c *WafClient) newRequest(op *request.Operation, params, data interface{}) *request.Request {
	req := c.NewRequest(op, params, data)

	return req
}

func (c *WafClient) UpdateDomain(input *waf.UpdateDomainInput) (*waf.UpdateDomainOutput, error) {
	req, out := c.UpdateDomainRequest(input)
	return out, req.Send()
}

func (c *WafClient) UpdateDomainRequest(input *waf.UpdateDomainInput) (req *request.Request, output *waf.UpdateDomainOutput) {
	op := &request.Operation{
		Name:       "UpdateDomain",
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &waf.UpdateDomainInput{}
	}

	output = &waf.UpdateDomainOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *WafClient) ListDomain(input *waf.ListDomainInput) (*waf.ListDomainOutput, error) {
	req, out := c.ListDomainRequest(input)
	return out, req.Send()
}

func (c *WafClient) ListDomainRequest(input *waf.ListDomainInput) (req *request.Request, output *waf.ListDomainOutput) {
	op := &request.Operation{
		Name:       "ListDomain",
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &waf.ListDomainInput{}
	}

	output = &waf.ListDomainOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}
