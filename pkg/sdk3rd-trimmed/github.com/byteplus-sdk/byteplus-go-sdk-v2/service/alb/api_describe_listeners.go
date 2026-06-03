package alb

import (
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/request"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/service/alb"
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

func (c *ALB) DescribeListenersWithContext(ctx byteplus.Context, input *DescribeListenersInput, opts ...request.Option) (*DescribeListenersOutput, error) {
	req, out := c.DescribeListenersRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type DescribeListenersInput = alb.DescribeListenersInput

type DescribeListenersOutput = alb.DescribeListenersOutput
