package clb

import (
	"github.com/volcengine/volcengine-go-sdk/service/clb"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opDescribeListenerAttributes = "DescribeListenerAttributes"

func (c *CLB) DescribeListenerAttributesRequest(input *DescribeListenerAttributesInput) (req *request.Request, output *DescribeListenerAttributesOutput) {
	op := &request.Operation{
		Name:       opDescribeListenerAttributes,
		HTTPMethod: "GET",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &DescribeListenerAttributesInput{}
	}

	output = &DescribeListenerAttributesOutput{}
	req = c.newRequest(op, input, output)

	return
}

func (c *CLB) DescribeListenerAttributesWithContext(ctx volcengine.Context, input *DescribeListenerAttributesInput, opts ...request.Option) (*DescribeListenerAttributesOutput, error) {
	req, out := c.DescribeListenerAttributesRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type DescribeListenerAttributesInput = clb.DescribeListenerAttributesInput

type DescribeListenerAttributesOutput = clb.DescribeListenerAttributesOutput

type DomainExtensionForDescribeListenerAttributesOutput = clb.DomainExtensionForDescribeListenerAttributesOutput
