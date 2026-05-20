package alb

import (
	"github.com/volcengine/volcengine-go-sdk/service/alb"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opModifyListenerAttributes = "ModifyListenerAttributes"

func (c *ALB) ModifyListenerAttributesRequest(input *ModifyListenerAttributesInput) (req *request.Request, output *ModifyListenerAttributesOutput) {
	op := &request.Operation{
		Name:       opModifyListenerAttributes,
		HTTPMethod: "GET",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &ModifyListenerAttributesInput{}
	}

	output = &ModifyListenerAttributesOutput{}
	req = c.newRequest(op, input, output)

	return
}

func (c *ALB) ModifyListenerAttributes(input *ModifyListenerAttributesInput) (*ModifyListenerAttributesOutput, error) {
	req, out := c.ModifyListenerAttributesRequest(input)
	return out, req.Send()
}

func (c *ALB) ModifyListenerAttributesWithContext(ctx volcengine.Context, input *ModifyListenerAttributesInput, opts ...request.Option) (*ModifyListenerAttributesOutput, error) {
	req, out := c.ModifyListenerAttributesRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type ModifyListenerAttributesInput = alb.ModifyListenerAttributesInput

type ModifyListenerAttributesOutput = alb.ModifyListenerAttributesOutput

type DomainExtensionForModifyListenerAttributesInput = alb.DomainExtensionForModifyListenerAttributesInput
