package clb

import (
	"github.com/volcengine/volcengine-go-sdk/service/clb"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
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

func (c *CLB) DescribeLoadBalancerAttributesWithContext(ctx volcengine.Context, input *DescribeLoadBalancerAttributesInput, opts ...request.Option) (*DescribeLoadBalancerAttributesOutput, error) {
	req, out := c.DescribeLoadBalancerAttributesRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type DescribeLoadBalancerAttributesInput = clb.DescribeLoadBalancerAttributesInput

type DescribeLoadBalancerAttributesOutput = clb.DescribeLoadBalancerAttributesOutput
