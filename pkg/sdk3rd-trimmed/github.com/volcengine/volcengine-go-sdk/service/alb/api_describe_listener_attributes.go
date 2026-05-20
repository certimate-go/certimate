package alb

import (
	"github.com/volcengine/volcengine-go-sdk/service/alb"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opDescribeListenerAttributes = "DescribeListenerAttributes"

func (c *ALB) DescribeListenerAttributesRequest(input *DescribeListenerAttributesInput) (req *request.Request, output *DescribeListenerAttributesOutput) {
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

func (c *ALB) DescribeListenerAttributesWithContext(ctx volcengine.Context, input *DescribeListenerAttributesInput, opts ...request.Option) (*DescribeListenerAttributesOutput, error) {
	req, out := c.DescribeListenerAttributesRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type DescribeListenerAttributesInput = alb.DescribeListenerAttributesInput

type DescribeListenerAttributesOutput = alb.DescribeListenerAttributesOutput

type DomainExtensionForDescribeListenerAttributesOutput = alb.DomainExtensionForDescribeListenerAttributesOutput
