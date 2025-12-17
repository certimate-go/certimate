package ulb

import (
	ucloudlb "github.com/ucloud/ucloud-sdk-go/services/ulb"
)

type DescribeSSLRequest = ucloudlb.DescribeSSLRequest

type DescribeSSLResponse = ucloudlb.DescribeSSLResponse

func (c *ULBClient) NewDescribeSSLRequest() *DescribeSSLRequest {
	req := &DescribeSSLRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	return req
}

func (c *ULBClient) DescribeSSL(req *DescribeSSLRequest) (*DescribeSSLResponse, error) {
	var err error
	var res DescribeSSLResponse

	reqCopier := *req

	err = c.Client.InvokeAction("DescribeSSL", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
