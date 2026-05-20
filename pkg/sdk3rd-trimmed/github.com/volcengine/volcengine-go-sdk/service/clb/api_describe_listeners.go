package clb

import (
	"github.com/volcengine/volcengine-go-sdk/service/clb"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opDescribeListeners = "DescribeListeners"

func (c *CLB) DescribeListenersRequest(input *DescribeListenersInput) (req *request.Request, output *DescribeListenersOutput) {
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

func (c *CLB) DescribeListenersWithContext(ctx volcengine.Context, input *DescribeListenersInput, opts ...request.Option) (*DescribeListenersOutput, error) {
	req, out := c.DescribeListenersRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type DescribeListenersInput = clb.DescribeListenersInput

type DescribeListenersOutput = clb.DescribeListenersOutput
