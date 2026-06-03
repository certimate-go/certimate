package apig

import (
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/request"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/service/apig"
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

func (c *APIG) GetCustomDomainWithContext(ctx byteplus.Context, input *GetCustomDomainInput, opts ...request.Option) (*GetCustomDomainOutput, error) {
	req, out := c.GetCustomDomainRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type GetCustomDomainInput = apig.GetCustomDomainInput

type GetCustomDomainOutput = apig.GetCustomDomainOutput
