package upathx

import (
	ucloudpathx "github.com/ucloud/ucloud-sdk-go/services/pathx"
)

type UnbindPathXSSLRequest = ucloudpathx.UnBindPathXSSLRequest

type UnbindPathXSSLResponse = ucloudpathx.UnBindPathXSSLResponse

func (c *UPathXClient) NewUnbindPathXSSLRequest() *UnbindPathXSSLRequest {
	req := &UnbindPathXSSLRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	return req
}

func (c *UPathXClient) UnbindPathXSSL(req *UnbindPathXSSLRequest) (*UnbindPathXSSLResponse, error) {
	var err error
	var res UnbindPathXSSLResponse

	reqCopier := *req

	err = c.Client.InvokeAction("UnBindPathXSSL", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
