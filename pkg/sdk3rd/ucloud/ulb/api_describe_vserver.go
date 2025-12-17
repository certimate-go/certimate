package ulb

import (
	ucloudlb "github.com/ucloud/ucloud-sdk-go/services/ulb"
)

type DescribeVServerRequest = ucloudlb.DescribeVServerRequest

type DescribeVServerResponse = ucloudlb.DescribeVServerResponse

func (c *ULBClient) NewDescribeVServerRequest() *DescribeVServerRequest {
	req := &DescribeVServerRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	return req
}

func (c *ULBClient) DescribeVServer(req *DescribeVServerRequest) (*DescribeVServerResponse, error) {
	var err error
	var res DescribeVServerResponse

	reqCopier := *req

	err = c.Client.InvokeAction("DescribeVServer", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
