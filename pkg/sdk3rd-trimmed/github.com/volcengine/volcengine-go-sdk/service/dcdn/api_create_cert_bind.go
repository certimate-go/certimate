package dcdn

import (
	"github.com/volcengine/volcengine-go-sdk/service/dcdn"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opCreateCertBind = "CreateCertBind"

func (c *DCDN) CreateCertBindRequest(input *CreateCertBindInput) (req *request.Request, output *CreateCertBindOutput) {
	op := &request.Operation{
		Name:       opCreateCertBind,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &CreateCertBindInput{}
	}

	output = &CreateCertBindOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *DCDN) CreateCertBindWithContext(ctx volcengine.Context, input *CreateCertBindInput, opts ...request.Option) (*CreateCertBindOutput, error) {
	req, out := c.CreateCertBindRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type CreateCertBindInput = dcdn.CreateCertBindInput

type CreateCertBindOutput = dcdn.CreateCertBindOutput
