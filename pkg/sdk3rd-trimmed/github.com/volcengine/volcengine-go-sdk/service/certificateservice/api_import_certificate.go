package certificateservice

import (
	"github.com/volcengine/volcengine-go-sdk/service/certificateservice"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opImportCertificate = "ImportCertificate"

func (c *CERTIFICATESERVICE) ImportCertificateRequest(input *ImportCertificateInput) (req *request.Request, output *ImportCertificateOutput) {
	op := &request.Operation{
		Name:       opImportCertificate,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &ImportCertificateInput{}
	}

	output = &ImportCertificateOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *CERTIFICATESERVICE) ImportCertificateWithContext(ctx volcengine.Context, input *ImportCertificateInput, opts ...request.Option) (*ImportCertificateOutput, error) {
	req, out := c.ImportCertificateRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type ImportCertificateInput = certificateservice.ImportCertificateInput

type ImportCertificateOutput = certificateservice.ImportCertificateOutput

type CertificateInfoForImportCertificateInput = certificateservice.CertificateInfoForImportCertificateInput
