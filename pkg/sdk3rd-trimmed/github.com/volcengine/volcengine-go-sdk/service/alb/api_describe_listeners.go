package alb

import (
	"github.com/volcengine/volcengine-go-sdk/service/alb"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opDescribeListeners = "DescribeListeners"

func (c *ALB) DescribeListenersRequest(input *DescribeListenersInput) (req *request.Request, output *DescribeListenersOutput) {
	op := &request.Operation{
		Name:       opDescribeListeners,
		HTTPMethod: "GET",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &DescribeListenersInput{}
	}

	output = &DescribeListenersOutput{}
	req = c.newRequest(op, input, output)

	return
}

func (c *ALB) DescribeListeners(input *DescribeListenersInput) (*DescribeListenersOutput, error) {
	req, out := c.DescribeListenersRequest(input)
	return out, req.Send()
}

func (c *ALB) DescribeListenersWithContext(ctx volcengine.Context, input *DescribeListenersInput, opts ...request.Option) (*DescribeListenersOutput, error) {
	req, out := c.DescribeListenersRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type DescribeListenersInput = alb.DescribeListenersInput

type DescribeListenersOutput = alb.DescribeListenersOutput
