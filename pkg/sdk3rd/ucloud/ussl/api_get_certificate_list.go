package ussl

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

type GetCertificateListRequest struct {
	request.CommonBase

	Mode           *string `required:"true"`
	StateCode      *string `required:"false"`
	Brand          *string `required:"false"`
	CaOrganization *string `required:"false"`
	Domain         *string `required:"false"`
	Sort           *string `required:"false"`
	Page           *int    `required:"false"`
	PageSize       *int    `required:"false"`
}

type GetCertificateListResponse struct {
	response.CommonBase

	CertificateList []*CertificateListItem
	TotalCount      int
}

func (c *USSLClient) NewGetCertificateListRequest() *GetCertificateListRequest {
	req := &GetCertificateListRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	req.SetEncoder(request.NewJSONEncoder(c.GetConfig(), c.GetCredential()))
	return req
}

func (c *USSLClient) GetCertificateList(req *GetCertificateListRequest) (*GetCertificateListResponse, error) {
	var err error
	var res GetCertificateListResponse

	reqCopier := *req

	err = c.Client.InvokeAction("GetCertificateList", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
