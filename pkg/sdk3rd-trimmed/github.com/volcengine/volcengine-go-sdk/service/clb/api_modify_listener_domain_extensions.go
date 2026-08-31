package clb

import (
	"github.com/volcengine/volcengine-go-sdk/service/clb"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opModifyListenerDomainExtensions = "ModifyListenerDomainExtensions"

func (c *CLB) ModifyListenerDomainExtensionsRequest(input *ModifyListenerDomainExtensionsInput) (req *request.Request, output *ModifyListenerDomainExtensionsOutput) {
	op := &request.Operation{
		Name:       opModifyListenerDomainExtensions,
		HTTPMethod: "GET",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &ModifyListenerDomainExtensionsInput{}
	}

	output = &ModifyListenerDomainExtensionsOutput{}
	req = c.newRequest(op, input, output)

	return
}

func (c *CLB) ModifyListenerDomainExtensionsWithContext(ctx volcengine.Context, input *ModifyListenerDomainExtensionsInput, opts ...request.Option) (*ModifyListenerDomainExtensionsOutput, error) {
	req, out := c.ModifyListenerDomainExtensionsRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type ModifyListenerDomainExtensionsInput = clb.ModifyListenerDomainExtensionsInput

type ModifyListenerDomainExtensionsOutput = clb.ModifyListenerDomainExtensionsOutput

type ModifyDomainExtensionForModifyListenerDomainExtensionsInput = clb.ModifyDomainExtensionForModifyListenerDomainExtensionsInput
