package alb

import (
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/request"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/service/alb"
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

func (c *ALB) ModifyListenerAttributesWithContext(ctx byteplus.Context, input *ModifyListenerAttributesInput, opts ...request.Option) (*ModifyListenerAttributesOutput, error) {
	req, out := c.ModifyListenerAttributesRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type ModifyListenerAttributesInput = alb.ModifyListenerAttributesInput

type ModifyListenerAttributesOutput = alb.ModifyListenerAttributesOutput

type DomainExtensionForModifyListenerAttributesInput = alb.DomainExtensionForModifyListenerAttributesInput
