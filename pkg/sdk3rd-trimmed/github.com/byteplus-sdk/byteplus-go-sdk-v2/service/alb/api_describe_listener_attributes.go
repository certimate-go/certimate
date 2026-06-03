package alb

import (
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/request"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/service/alb"
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

func (c *ALB) DescribeListenerAttributesWithContext(ctx byteplus.Context, input *DescribeListenerAttributesInput, opts ...request.Option) (*DescribeListenerAttributesOutput, error) {
	req, out := c.DescribeListenerAttributesRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type DescribeListenerAttributesInput = alb.DescribeListenerAttributesInput

type DescribeListenerAttributesOutput = alb.DescribeListenerAttributesOutput

type DomainExtensionForDescribeListenerAttributesOutput = alb.DomainExtensionForDescribeListenerAttributesOutput
