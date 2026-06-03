package certificateservice

import (
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/byteplusutil"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/request"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/response"
)

const opUploadCertificate = "UploadCertificate"

func (c *CERTIFICATESERVICE) UploadCertificateRequest(input *UploadCertificateInput) (req *request.Request, output *UploadCertificateOutput) {
	op := &request.Operation{
		Name:       opUploadCertificate,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &UploadCertificateInput{}
	}

	output = &UploadCertificateOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *CERTIFICATESERVICE) UploadCertificateWithContext(ctx byteplus.Context, input *UploadCertificateInput, opts ...request.Option) (*UploadCertificateOutput, error) {
	req, out := c.UploadCertificateRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type UploadCertificateInput struct {
	_ struct{} `type:"structure" json:",omitempty"`

	CertificateInfo *CertificateInfoForUploadCertificateInput `type:"structure" json:",omitempty"`

	NoVerifyAndFixChain *bool `type:"boolean" json:",omitempty"`

	ProjectName *string `type:"string" json:",omitempty"`

	Repeatable *bool `type:"boolean" json:",omitempty"`

	Tag *string `type:"string" json:",omitempty"`

	Tags []*TagForUploadCertificateInput `type:"list" json:",omitempty"`
}

func (s UploadCertificateInput) String() string {
	return byteplusutil.Prettify(s)
}

func (s UploadCertificateInput) GoString() string {
	return s.String()
}

type UploadCertificateOutput struct {
	_ struct{} `type:"structure" json:",omitempty"`

	Metadata *response.ResponseMetadata

	InstanceId *string `type:"string" json:",omitempty"`

	RepeatId *string `type:"string" json:",omitempty"`
}

func (s UploadCertificateOutput) String() string {
	return byteplusutil.Prettify(s)
}

func (s UploadCertificateOutput) GoString() string {
	return s.String()
}

type CertificateInfoForUploadCertificateInput struct {
	_ struct{} `type:"structure" json:",omitempty"`

	CertificateChain *string `type:"string" json:",omitempty"`

	PrivateKey *string `type:"string" json:",omitempty"`
}

func (s CertificateInfoForUploadCertificateInput) String() string {
	return byteplusutil.Prettify(s)
}

func (s CertificateInfoForUploadCertificateInput) GoString() string {
	return s.String()
}

type TagForUploadCertificateInput struct {
	_ struct{} `type:"structure" json:",omitempty"`

	Key *string `type:"string" json:",omitempty"`

	Value *string `type:"string" json:",omitempty"`
}

func (s TagForUploadCertificateInput) String() string {
	return byteplusutil.Prettify(s)
}

func (s TagForUploadCertificateInput) GoString() string {
	return s.String()
}
