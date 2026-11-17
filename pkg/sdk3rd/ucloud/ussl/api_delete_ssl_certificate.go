package ussl

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

type DeleteSSLCertificateRequest struct {
	request.CommonBase

	CertificateMode *string `required:"true"`
	CertificateID   *int    `required:"true"`
}

type DeleteSSLCertificateResponse struct {
	response.CommonBase
}

func (c *USSLClient) NewDeleteSSLCertificateRequest() *DeleteSSLCertificateRequest {
	req := &DeleteSSLCertificateRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	req.SetEncoder(request.NewJSONEncoder(c.GetConfig(), c.GetCredential()))
	return req
}

func (c *USSLClient) DeleteSSLCertificate(req *DeleteSSLCertificateRequest) (*DeleteSSLCertificateResponse, error) {
	var err error
	var res DeleteSSLCertificateResponse

	reqCopier := *req

	err = c.Client.InvokeAction("DeleteSSLCertificate", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
