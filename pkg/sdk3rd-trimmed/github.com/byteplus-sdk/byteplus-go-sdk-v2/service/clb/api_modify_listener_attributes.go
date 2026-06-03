package clb

import (
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/request"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/service/clb"
)

const opModifyListenerAttributes = "ModifyListenerAttributes"

func (c *CLB) ModifyListenerAttributesRequest(input *ModifyListenerAttributesInput) (req *request.Request, output *ModifyListenerAttributesOutput) {
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

func (c *CLB) ModifyListenerAttributesWithContext(ctx byteplus.Context, input *ModifyListenerAttributesInput, opts ...request.Option) (*ModifyListenerAttributesOutput, error) {
	req, out := c.ModifyListenerAttributesRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type ModifyListenerAttributesInput = clb.ModifyListenerAttributesInput

type ModifyListenerAttributesOutput = clb.ModifyListenerAttributesOutput
