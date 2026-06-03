package apig

import (
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/request"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/service/apig"
)

const opListCustomDomains = "ListCustomDomains"

func (c *APIG) ListCustomDomainsRequest(input *ListCustomDomainsInput) (req *request.Request, output *ListCustomDomainsOutput) {
	op := &request.Operation{
		Name:       opListCustomDomains,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &ListCustomDomainsInput{}
	}

	output = &ListCustomDomainsOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *APIG) ListCustomDomainsWithContext(ctx byteplus.Context, input *ListCustomDomainsInput, opts ...request.Option) (*ListCustomDomainsOutput, error) {
	req, out := c.ListCustomDomainsRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type ListCustomDomainsInput = apig.ListCustomDomainsInput

type ListCustomDomainsOutput = apig.ListCustomDomainsOutput

type ItemForListCustomDomainsOutput = apig.ItemForListCustomDomainsOutput
