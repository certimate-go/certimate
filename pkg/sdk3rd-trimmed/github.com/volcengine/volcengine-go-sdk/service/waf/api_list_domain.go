package waf

import (
	"github.com/volcengine/volcengine-go-sdk/service/waf"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opListDomain = "ListDomain"

func (c *WAF) ListDomainRequest(input *ListDomainInput) (req *request.Request, output *ListDomainOutput) {
	op := &request.Operation{
		Name:       opListDomain,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &ListDomainInput{}
	}

	output = &ListDomainOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *WAF) ListDomainWithContext(ctx volcengine.Context, input *ListDomainInput, opts ...request.Option) (*ListDomainOutput, error) {
	req, out := c.ListDomainRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type ListDomainInput = waf.ListDomainInput

type ListDomainOutput = waf.ListDomainOutput
