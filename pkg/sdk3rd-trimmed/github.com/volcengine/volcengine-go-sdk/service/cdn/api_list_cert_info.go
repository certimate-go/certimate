package cdn

import (
	"github.com/volcengine/volcengine-go-sdk/service/cdn"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opListCertInfo = "ListCertInfo"

func (c *CDN) ListCertInfoRequest(input *ListCertInfoInput) (req *request.Request, output *ListCertInfoOutput) {
	op := &request.Operation{
		Name:       opListCertInfo,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &ListCertInfoInput{}
	}

	output = &ListCertInfoOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *CDN) ListCertInfoWithContext(ctx volcengine.Context, input *ListCertInfoInput, opts ...request.Option) (*ListCertInfoOutput, error) {
	req, out := c.ListCertInfoRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type ListCertInfoInput = cdn.ListCertInfoInput

type ListCertInfoOutput = cdn.ListCertInfoOutput
