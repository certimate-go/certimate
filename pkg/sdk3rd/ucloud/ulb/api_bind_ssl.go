package ulb

import (
	ucloudlb "github.com/ucloud/ucloud-sdk-go/services/ulb"
)

type BindSSLRequest = ucloudlb.BindSSLRequest

type BindSSLResponse = ucloudlb.BindSSLResponse

func (c *ULBClient) NewBindSSLRequest() *BindSSLRequest {
	req := &BindSSLRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	return req
}

func (c *ULBClient) BindSSL(req *BindSSLRequest) (*BindSSLResponse, error) {
	var err error
	var res BindSSLResponse

	reqCopier := *req

	err = c.Client.InvokeAction("BindSSL", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
