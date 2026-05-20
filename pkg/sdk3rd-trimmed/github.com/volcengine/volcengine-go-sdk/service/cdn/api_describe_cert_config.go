package cdn

import (
	"github.com/volcengine/volcengine-go-sdk/service/cdn"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opDescribeCertConfig = "DescribeCertConfig"

func (c *CDN) DescribeCertConfigRequest(input *DescribeCertConfigInput) (req *request.Request, output *DescribeCertConfigOutput) {
	op := &request.Operation{
		Name:       opDescribeCertConfig,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &DescribeCertConfigInput{}
	}

	output = &DescribeCertConfigOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *CDN) DescribeCertConfigWithContext(ctx volcengine.Context, input *DescribeCertConfigInput, opts ...request.Option) (*DescribeCertConfigOutput, error) {
	req, out := c.DescribeCertConfigRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type DescribeCertConfigInput = cdn.DescribeCertConfigInput

type DescribeCertConfigOutput = cdn.DescribeCertConfigOutput
