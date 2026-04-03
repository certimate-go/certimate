package internal

import (
	"github.com/volcengine/volcengine-go-sdk/service/apig"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/client"
	"github.com/volcengine/volcengine-go-sdk/volcengine/client/metadata"
	"github.com/volcengine/volcengine-go-sdk/volcengine/corehandlers"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
	"github.com/volcengine/volcengine-go-sdk/volcengine/signer/volc"
	"github.com/volcengine/volcengine-go-sdk/volcengine/volcenginequery"
)

// This is a partial copy of https://github.com/volcengine/volcengine-go-sdk/blob/master/service/apig/service_apig.go
// to lightweight the vendor packages in the built binary.
type ApigClient struct {
	*client.Client
}

func NewApigClient(p client.ConfigProvider, cfgs ...*volcengine.Config) *ApigClient {
	c := p.ClientConfig(apig.EndpointsID, cfgs...)
	return newApigClient(*c.Config, c.Handlers, c.Endpoint, c.SigningRegion, c.SigningName)
}

func newApigClient(cfg volcengine.Config, handlers request.Handlers, endpoint, signingRegion, signingName string) *ApigClient {
	svc := &ApigClient{
		Client: client.New(
			cfg,
			metadata.ClientInfo{
				ServiceName:   apig.ServiceName,
				ServiceID:     apig.ServiceID,
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
	svc.Handlers.Sign.PushBackNamed(volc.SignRequestHandler)
	svc.Handlers.Build.PushBackNamed(volcenginequery.BuildHandler)
	svc.Handlers.Unmarshal.PushBackNamed(volcenginequery.UnmarshalHandler)
	svc.Handlers.UnmarshalMeta.PushBackNamed(volcenginequery.UnmarshalMetaHandler)
	svc.Handlers.UnmarshalError.PushBackNamed(volcenginequery.UnmarshalErrorHandler)

	return svc
}

func (c *ApigClient) newRequest(op *request.Operation, params, data interface{}) *request.Request {
	req := c.NewRequest(op, params, data)

	return req
}

func (c *ApigClient) GetCustomDomainRequest(input *apig.GetCustomDomainInput) (req *request.Request, output *apig.GetCustomDomainOutput) {
	op := &request.Operation{
		Name:       "GetCustomDomain",
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &apig.GetCustomDomainInput{}
	}

	output = &apig.GetCustomDomainOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *ApigClient) GetCustomDomain(input *apig.GetCustomDomainInput) (*apig.GetCustomDomainOutput, error) {
	req, out := c.GetCustomDomainRequest(input)
	return out, req.Send()
}

func (c *ApigClient) ListCustomDomainsRequest(input *apig.ListCustomDomainsInput) (req *request.Request, output *apig.ListCustomDomainsOutput) {
	op := &request.Operation{
		Name:       "ListCustomDomains",
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &apig.ListCustomDomainsInput{}
	}

	output = &apig.ListCustomDomainsOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *ApigClient) ListCustomDomains(input *apig.ListCustomDomainsInput) (*apig.ListCustomDomainsOutput, error) {
	req, out := c.ListCustomDomainsRequest(input)
	return out, req.Send()
}

func (c *ApigClient) UpdateCustomDomainRequest(input *apig.UpdateCustomDomainInput) (req *request.Request, output *apig.UpdateCustomDomainOutput) {
	op := &request.Operation{
		Name:       "UpdateCustomDomain",
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &apig.UpdateCustomDomainInput{}
	}

	output = &apig.UpdateCustomDomainOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *ApigClient) UpdateCustomDomain(input *apig.UpdateCustomDomainInput) (*apig.UpdateCustomDomainOutput, error) {
	req, out := c.UpdateCustomDomainRequest(input)
	return out, req.Send()
}
