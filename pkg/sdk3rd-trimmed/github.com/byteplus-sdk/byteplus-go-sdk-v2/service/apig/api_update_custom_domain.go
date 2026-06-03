package apig

import (
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/request"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/service/apig"
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

func (c *APIG) UpdateCustomDomainWithContext(ctx byteplus.Context, input *UpdateCustomDomainInput, opts ...request.Option) (*UpdateCustomDomainOutput, error) {
	req, out := c.UpdateCustomDomainRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type UpdateCustomDomainInput = apig.UpdateCustomDomainInput

type UpdateCustomDomainOutput = apig.UpdateCustomDomainOutput
