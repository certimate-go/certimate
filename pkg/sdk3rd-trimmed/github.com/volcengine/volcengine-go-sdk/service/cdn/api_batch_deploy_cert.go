package cdn

import (
	"github.com/volcengine/volcengine-go-sdk/service/cdn"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opBatchDeployCert = "BatchDeployCert"

func (c *CDN) BatchDeployCertRequest(input *BatchDeployCertInput) (req *request.Request, output *BatchDeployCertOutput) {
	op := &request.Operation{
		Name:       opBatchDeployCert,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &BatchDeployCertInput{}
	}

	output = &BatchDeployCertOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *CDN) BatchDeployCertWithContext(ctx volcengine.Context, input *BatchDeployCertInput, opts ...request.Option) (*BatchDeployCertOutput, error) {
	req, out := c.BatchDeployCertRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type BatchDeployCertInput = cdn.BatchDeployCertInput

type BatchDeployCertOutput = cdn.BatchDeployCertOutput
