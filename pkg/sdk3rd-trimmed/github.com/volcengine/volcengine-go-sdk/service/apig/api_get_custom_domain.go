package apig

import (
	"github.com/volcengine/volcengine-go-sdk/service/apig"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opGetCustomDomain = "GetCustomDomain"

func (c *APIG) GetCustomDomainRequest(input *GetCustomDomainInput) (req *request.Request, output *GetCustomDomainOutput) {
	op := &request.Operation{
		Name:       opGetCustomDomain,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &GetCustomDomainInput{}
	}

	output = &GetCustomDomainOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *APIG) GetCustomDomainWithContext(ctx volcengine.Context, input *GetCustomDomainInput, opts ...request.Option) (*GetCustomDomainOutput, error) {
	req, out := c.GetCustomDomainRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type GetCustomDomainInput = apig.GetCustomDomainInput

type GetCustomDomainOutput = apig.GetCustomDomainOutput
