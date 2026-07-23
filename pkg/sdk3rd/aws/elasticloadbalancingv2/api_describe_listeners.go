package elasticloadbalancingv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
)

type DescribeListenersRequest = elasticloadbalancingv2.DescribeListenersInput

type DescribeListenersResponse = elasticloadbalancingv2.DescribeListenersOutput

func (c *Client) DescribeListeners(req *DescribeListenersRequest) (*DescribeListenersResponse, error) {
	return c.DescribeListenersWithContext(context.Background(), req)
}

func (c *Client) DescribeListenersWithContext(ctx context.Context, req *DescribeListenersRequest) (*DescribeListenersResponse, error) {
	params := &struct {
		DescribeListenersRequest `json:",inline"`
		Action                   string
		Version                  string
	}{
		DescribeListenersRequest: *req,
		Action:                   "DescribeListeners",
		Version:                  "2015-12-01",
	}

	httpreq, err := c.newRequest(params)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &DescribeListenersResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
