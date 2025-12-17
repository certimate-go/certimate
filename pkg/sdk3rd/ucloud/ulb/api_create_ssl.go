package ulb

import (
	ucloudlb "github.com/ucloud/ucloud-sdk-go/services/ulb"
)

type CreateSSLRequest = ucloudlb.CreateSSLRequest

type CreateSSLResponse = ucloudlb.CreateSSLResponse

func (c *ULBClient) NewCreateSSLRequest() *CreateSSLRequest {
	req := &CreateSSLRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	return req
}

func (c *ULBClient) CreateSSL(req *CreateSSLRequest) (*CreateSSLResponse, error) {
	var err error
	var res CreateSSLResponse

	reqCopier := *req

	err = c.Client.InvokeAction("CreateSSL", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
