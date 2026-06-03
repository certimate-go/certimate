package clb

import (
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/request"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/service/clb"
)

const opDescribeLoadBalancerAttributes = "DescribeLoadBalancerAttributes"

func (c *CLB) DescribeLoadBalancerAttributesRequest(input *DescribeLoadBalancerAttributesInput) (req *request.Request, output *DescribeLoadBalancerAttributesOutput) {
	op := &request.Operation{
		Name:       opDescribeLoadBalancerAttributes,
		HTTPMethod: "GET",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &DescribeLoadBalancerAttributesInput{}
	}

	output = &DescribeLoadBalancerAttributesOutput{}
	req = c.newRequest(op, input, output)

	return
}

func (c *CLB) DescribeLoadBalancerAttributesWithContext(ctx byteplus.Context, input *DescribeLoadBalancerAttributesInput, opts ...request.Option) (*DescribeLoadBalancerAttributesOutput, error) {
	req, out := c.DescribeLoadBalancerAttributesRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type DescribeLoadBalancerAttributesInput = clb.DescribeLoadBalancerAttributesInput

type DescribeLoadBalancerAttributesOutput = clb.DescribeLoadBalancerAttributesOutput
