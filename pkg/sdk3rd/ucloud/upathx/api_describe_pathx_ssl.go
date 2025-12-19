package upathx

import (
	ucloudpathx "github.com/ucloud/ucloud-sdk-go/services/pathx"
)

type DescribePathXSSLRequest = ucloudpathx.DescribePathXSSLRequest

type DescribePathXSSLResponse = ucloudpathx.DescribePathXSSLResponse

func (c *UPathXClient) NewDescribePathXSSLRequest() *DescribePathXSSLRequest {
	req := &DescribePathXSSLRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	return req
}

func (c *UPathXClient) DescribePathXSSL(req *DescribePathXSSLRequest) (*DescribePathXSSLResponse, error) {
	var err error
	var res DescribePathXSSLResponse

	reqCopier := *req

	err = c.Client.InvokeAction("DescribePathXSSL", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
