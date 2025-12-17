package ulb

import (
	ucloudlb "github.com/ucloud/ucloud-sdk-go/services/ulb"
)

type DescribeListenersRequest = ucloudlb.DescribeListenersRequest

type DescribeListenersResponse = ucloudlb.DescribeListenersResponse

func (c *ULBClient) NewDescribeListenersRequest() *DescribeListenersRequest {
	req := &DescribeListenersRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	return req
}

func (c *ULBClient) DescribeListeners(req *DescribeListenersRequest) (*DescribeListenersResponse, error) {
	var err error
	var res DescribeListenersResponse

	reqCopier := *req

	err = c.Client.InvokeAction("DescribeListeners", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
