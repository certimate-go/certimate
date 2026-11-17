package ussl

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

type UploadNormalCertificateRequest struct {
	request.CommonBase

	CertificateName *string `required:"true"`
	SslPublicKey    *string `required:"true"`
	SslPrivateKey   *string `required:"true"`
	SslMD5          *string `required:"true"`
	SslCaKey        *string `required:"false"`
}

type UploadNormalCertificateResponse struct {
	response.CommonBase

	CertificateID  int
	LongResourceID string
}

func (c *USSLClient) NewUploadNormalCertificateRequest() *UploadNormalCertificateRequest {
	req := &UploadNormalCertificateRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(false)
	req.SetEncoder(request.NewJSONEncoder(c.GetConfig(), c.GetCredential()))
	return req
}

func (c *USSLClient) UploadNormalCertificate(req *UploadNormalCertificateRequest) (*UploadNormalCertificateResponse, error) {
	var err error
	var res UploadNormalCertificateResponse

	reqCopier := *req

	err = c.Client.InvokeAction("UploadNormalCertificate", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
