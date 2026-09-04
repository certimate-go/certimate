package certificateservice

import (
	"github.com/volcengine/volcengine-go-sdk/service/certificateservice"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opCertificateGetInstanceList = "CertificateGetInstanceList"

func (c *CERTIFICATESERVICE) CertificateGetInstanceListRequest(input *CertificateGetInstanceListInput) (req *request.Request, output *CertificateGetInstanceListOutput) {
	op := &request.Operation{
		Name:       opCertificateGetInstanceList,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &CertificateGetInstanceListInput{}
	}

	output = &CertificateGetInstanceListOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *CERTIFICATESERVICE) CertificateGetInstanceListWithContext(ctx volcengine.Context, input *CertificateGetInstanceListInput, opts ...request.Option) (*CertificateGetInstanceListOutput, error) {
	req, out := c.CertificateGetInstanceListRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type CertificateGetInstanceListInput = certificateservice.CertificateGetInstanceListInput

type CertificateGetInstanceListOutput = certificateservice.CertificateGetInstanceListOutput
