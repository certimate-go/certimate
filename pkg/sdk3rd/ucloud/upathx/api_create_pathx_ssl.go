package upathx

import (
	ucloudpathx "github.com/ucloud/ucloud-sdk-go/services/pathx"
)

type CreatePathXSSLRequest = ucloudpathx.CreatePathXSSLRequest

type CreatePathXSSLResponse = ucloudpathx.CreatePathXSSLResponse

func (c *UPathXClient) NewCreatePathXSSLRequest() *CreatePathXSSLRequest {
	req := &CreatePathXSSLRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	return req
}

func (c *UPathXClient) CreatePathXSSL(req *CreatePathXSSLRequest) (*CreatePathXSSLResponse, error) {
	var err error
	var res CreatePathXSSLResponse

	reqCopier := *req

	err = c.Client.InvokeAction("CreatePathXSSL", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
