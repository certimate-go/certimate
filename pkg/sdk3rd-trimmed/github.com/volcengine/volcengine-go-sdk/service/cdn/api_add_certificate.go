package cdn

import (
	"github.com/volcengine/volcengine-go-sdk/service/cdn"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opAddCertificate = "AddCertificate"

func (c *CDN) AddCertificateRequest(input *AddCertificateInput) (req *request.Request, output *AddCertificateOutput) {
	op := &request.Operation{
		Name:       opAddCertificate,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &AddCertificateInput{}
	}

	output = &AddCertificateOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *CDN) AddCertificateWithContext(ctx volcengine.Context, input *AddCertificateInput, opts ...request.Option) (*AddCertificateOutput, error) {
	req, out := c.AddCertificateRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type AddCertificateInput = cdn.AddCertificateInput

type AddCertificateOutput = cdn.AddCertificateOutput
