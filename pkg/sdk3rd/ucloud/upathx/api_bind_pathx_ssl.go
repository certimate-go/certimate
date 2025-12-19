package upathx

import (
	ucloudpathx "github.com/ucloud/ucloud-sdk-go/services/pathx"
)

type BindPathXSSLRequest = ucloudpathx.BindPathXSSLRequest

type BindPathXSSLResponse = ucloudpathx.BindPathXSSLResponse

func (c *UPathXClient) NewBindPathXSSLRequest() *BindPathXSSLRequest {
	req := &BindPathXSSLRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	return req
}

func (c *UPathXClient) BindPathXSSL(req *BindPathXSSLRequest) (*BindPathXSSLResponse, error) {
	var err error
	var res BindPathXSSLResponse

	reqCopier := *req

	err = c.Client.InvokeAction("BindPathXSSL", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
