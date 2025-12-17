package ulb

import (
	ucloudlb "github.com/ucloud/ucloud-sdk-go/services/ulb"
)

type UnbindSSLRequest = ucloudlb.UnbindSSLRequest

type UnbindSSLResponse = ucloudlb.UnbindSSLResponse

func (c *ULBClient) NewUnbindSSLRequest() *UnbindSSLRequest {
	req := &UnbindSSLRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	return req
}

func (c *ULBClient) UnbindSSL(req *UnbindSSLRequest) (*UnbindSSLResponse, error) {
	var err error
	var res UnbindSSLResponse

	reqCopier := *req

	err = c.Client.InvokeAction("UnbindSSL", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
