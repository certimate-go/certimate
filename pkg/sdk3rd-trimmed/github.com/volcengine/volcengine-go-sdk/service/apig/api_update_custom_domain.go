package apig

import (
	"github.com/volcengine/volcengine-go-sdk/service/apig"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opUpdateCustomDomain = "UpdateCustomDomain"

func (c *APIG) UpdateCustomDomainRequest(input *UpdateCustomDomainInput) (req *request.Request, output *UpdateCustomDomainOutput) {
	op := &request.Operation{
		Name:       opUpdateCustomDomain,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &UpdateCustomDomainInput{}
	}

	output = &UpdateCustomDomainOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *APIG) UpdateCustomDomainWithContext(ctx volcengine.Context, input *UpdateCustomDomainInput, opts ...request.Option) (*UpdateCustomDomainOutput, error) {
	req, out := c.UpdateCustomDomainRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type UpdateCustomDomainInput = apig.UpdateCustomDomainInput

type UpdateCustomDomainOutput = apig.UpdateCustomDomainOutput
