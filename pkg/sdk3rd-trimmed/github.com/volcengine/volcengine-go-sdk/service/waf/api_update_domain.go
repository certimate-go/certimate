package waf

import (
	"github.com/volcengine/volcengine-go-sdk/service/waf"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opUpdateDomain = "UpdateDomain"

func (c *WAF) UpdateDomainRequest(input *UpdateDomainInput) (req *request.Request, output *UpdateDomainOutput) {
	op := &request.Operation{
		Name:       opUpdateDomain,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &UpdateDomainInput{}
	}

	output = &UpdateDomainOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *WAF) UpdateDomainWithContext(ctx volcengine.Context, input *UpdateDomainInput, opts ...request.Option) (*UpdateDomainOutput, error) {
	req, out := c.UpdateDomainRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type UpdateDomainInput = waf.UpdateDomainInput

type UpdateDomainOutput = waf.UpdateDomainOutput

type ProtocolPortsForUpdateDomainInput = waf.ProtocolPortsForUpdateDomainInput
