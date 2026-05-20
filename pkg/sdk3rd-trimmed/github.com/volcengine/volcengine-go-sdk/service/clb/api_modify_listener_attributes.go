package clb

import (
	"github.com/volcengine/volcengine-go-sdk/service/clb"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
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

func (c *CLB) ModifyListenerAttributesWithContext(ctx volcengine.Context, input *ModifyListenerAttributesInput, opts ...request.Option) (*ModifyListenerAttributesOutput, error) {
	req, out := c.ModifyListenerAttributesRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type ModifyListenerAttributesInput = clb.ModifyListenerAttributesInput

type ModifyListenerAttributesOutput = clb.ModifyListenerAttributesOutput
