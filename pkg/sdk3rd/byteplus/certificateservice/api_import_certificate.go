package certificateservice

import (
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/byteplusutil"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/request"

	"github.com/certimate-go/certimate/pkg/sdk3rd-patched/github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/response"
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

func (c *CERTIFICATESERVICE) ImportCertificateWithContext(ctx byteplus.Context, input *ImportCertificateInput, opts ...request.Option) (*ImportCertificateOutput, error) {
	req, out := c.ImportCertificateRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type ImportCertificateInput struct {
	_ struct{} `type:"structure"`

	CertificateInfo *CertificateInfoForImportCertificateInput `type:"structure" json:",omitempty"`

	NoVerifyAndFixChain *bool `type:"boolean" json:",omitempty"`

	ProjectName *string `type:"string" json:",omitempty"`

	Repeatable *bool `type:"boolean" json:",omitempty"`

	Tag *string `type:"string" json:",omitempty"`

	Tags []*TagForImportCertificateInput `type:"list" json:",omitempty"`
}

func (s ImportCertificateInput) String() string {
	return byteplusutil.Prettify(s)
}

func (s ImportCertificateInput) GoString() string {
	return s.String()
}

type ImportCertificateOutput struct {
	_ struct{} `type:"structure"`

	Metadata *response.ResponseMetadata

	InstanceId *string `type:"string" json:",omitempty"`

	RepeatId *string `type:"string" json:",omitempty"`
}

func (s ImportCertificateOutput) String() string {
	return byteplusutil.Prettify(s)
}

func (s ImportCertificateOutput) GoString() string {
	return s.String()
}

type CertificateInfoForImportCertificateInput struct {
	_ struct{} `type:"structure"`

	CertificateChain *string `type:"string" json:",omitempty"`

	PrivateKey *string `type:"string" json:",omitempty"`
}

func (s CertificateInfoForImportCertificateInput) String() string {
	return byteplusutil.Prettify(s)
}

func (s CertificateInfoForImportCertificateInput) GoString() string {
	return s.String()
}

type TagForImportCertificateInput struct {
	_ struct{} `type:"structure"`

	Key *string `type:"string" json:",omitempty"`

	Value *string `type:"string" json:",omitempty"`
}

func (s TagForImportCertificateInput) String() string {
	return byteplusutil.Prettify(s)
}

func (s TagForImportCertificateInput) GoString() string {
	return s.String()
}
