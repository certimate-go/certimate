package certificateservice

import (
	"github.com/volcengine/volcengine-go-sdk/service/certificateservice"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opCertificateDeleteInstance = "CertificateDeleteInstance"

func (c *CERTIFICATESERVICE) CertificateDeleteInstanceRequest(input *CertificateDeleteInstanceInput) (req *request.Request, output *CertificateDeleteInstanceOutput) {
	op := &request.Operation{
		Name:       opCertificateDeleteInstance,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &CertificateDeleteInstanceInput{}
	}

	output = &CertificateDeleteInstanceOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *CERTIFICATESERVICE) CertificateDeleteInstanceWithContext(ctx volcengine.Context, input *CertificateDeleteInstanceInput, opts ...request.Option) (*CertificateDeleteInstanceOutput, error) {
	req, out := c.CertificateDeleteInstanceRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type CertificateDeleteInstanceInput = certificateservice.CertificateDeleteInstanceInput

type CertificateDeleteInstanceOutput = certificateservice.CertificateDeleteInstanceOutput
