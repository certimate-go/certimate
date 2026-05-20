package dcdn

import (
	"github.com/volcengine/volcengine-go-sdk/service/dcdn"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opListDomainConfig = "ListDomainConfig"

func (c *DCDN) ListDomainConfigRequest(input *ListDomainConfigInput) (req *request.Request, output *ListDomainConfigOutput) {
	op := &request.Operation{
		Name:       opListDomainConfig,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &ListDomainConfigInput{}
	}

	output = &ListDomainConfigOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *DCDN) ListDomainConfigWithContext(ctx volcengine.Context, input *ListDomainConfigInput, opts ...request.Option) (*ListDomainConfigOutput, error) {
	req, out := c.ListDomainConfigRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type ListDomainConfigInput = dcdn.ListDomainConfigInput

type ListDomainConfigOutput = dcdn.ListDomainConfigOutput
