package ussl

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

type GetCertificateDetailInfoRequest struct {
	request.CommonBase

	CertificateID *int `required:"true"`
}

type GetCertificateDetailInfoResponse struct {
	response.CommonBase

	CertificateInfo *CertificateInfo
}

func (c *USSLClient) NewGetCertificateDetailInfoRequest() *GetCertificateDetailInfoRequest {
	req := &GetCertificateDetailInfoRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	req.SetEncoder(request.NewJSONEncoder(c.GetConfig(), c.GetCredential()))
	return req
}

func (c *USSLClient) GetCertificateDetailInfo(req *GetCertificateDetailInfoRequest) (*GetCertificateDetailInfoResponse, error) {
	var err error
	var res GetCertificateDetailInfoResponse

	reqCopier := *req

	err = c.Client.InvokeAction("GetCertificateDetailInfo", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
