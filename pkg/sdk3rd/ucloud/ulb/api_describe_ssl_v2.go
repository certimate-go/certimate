package ulb

import (
	ucloudlb "github.com/ucloud/ucloud-sdk-go/services/ulb"
)

type DescribeSSLV2Request = ucloudlb.DescribeSSLV2Request

type DescribeSSLV2Response = ucloudlb.DescribeSSLV2Response

func (c *ULBClient) NewDescribeSSLV2Request() *DescribeSSLV2Request {
	req := &DescribeSSLV2Request{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	return req
}

func (c *ULBClient) DescribeSSLV2(req *DescribeSSLV2Request) (*DescribeSSLV2Response, error) {
	var err error
	var res DescribeSSLV2Response

	reqCopier := *req

	err = c.Client.InvokeAction("DescribeSSLV2", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
