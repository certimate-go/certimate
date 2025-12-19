package uewaf

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

type AddWafDomainCertificateInfoRequest struct {
	request.CommonBase

	Domain          *string `required:"true"`
	CertificateName *string `required:"true"`
	SslPublicKey    *string `required:"true"`
	SslPrivateKey   *string `required:"false"`
	SslMD           *string `required:"false"`
	SslKeyLess      *string `required:"false"`
}

type AddWafDomainCertificateInfoResponse struct {
	response.CommonBase

	Id int
}

func (c *UEWAFClient) NewAddWafDomainCertificateInfoRequest() *AddWafDomainCertificateInfoRequest {
	req := &AddWafDomainCertificateInfoRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	return req
}

func (c *UEWAFClient) AddWafDomainCertificateInfo(req *AddWafDomainCertificateInfoRequest) (*AddWafDomainCertificateInfoResponse, error) {
	var err error
	var res AddWafDomainCertificateInfoResponse

	reqCopier := *req

	err = c.Client.InvokeAction("AddWafDomainCertificateInfo", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
