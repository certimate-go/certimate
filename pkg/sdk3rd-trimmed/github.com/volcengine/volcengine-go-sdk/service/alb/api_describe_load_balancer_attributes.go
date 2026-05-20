package alb

import (
	"github.com/volcengine/volcengine-go-sdk/service/alb"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opDescribeLoadBalancerAttributes = "DescribeLoadBalancerAttributes"

func (c *ALB) DescribeLoadBalancerAttributesRequest(input *DescribeLoadBalancerAttributesInput) (req *request.Request, output *DescribeLoadBalancerAttributesOutput) {
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

func (c *ALB) DescribeLoadBalancerAttributes(input *DescribeLoadBalancerAttributesInput) (*DescribeLoadBalancerAttributesOutput, error) {
	req, out := c.DescribeLoadBalancerAttributesRequest(input)
	return out, req.Send()
}

func (c *ALB) DescribeLoadBalancerAttributesWithContext(ctx volcengine.Context, input *DescribeLoadBalancerAttributesInput, opts ...request.Option) (*DescribeLoadBalancerAttributesOutput, error) {
	req, out := c.DescribeLoadBalancerAttributesRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type DescribeLoadBalancerAttributesInput = alb.DescribeLoadBalancerAttributesInput

type DescribeLoadBalancerAttributesOutput = alb.DescribeLoadBalancerAttributesOutput
