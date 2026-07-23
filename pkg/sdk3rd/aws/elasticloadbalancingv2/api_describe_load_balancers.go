package elasticloadbalancingv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
)

type DescribeLoadBalancersRequest = elasticloadbalancingv2.DescribeLoadBalancersInput

type DescribeLoadBalancersResponse = elasticloadbalancingv2.DescribeLoadBalancersOutput

func (c *Client) DescribeLoadBalancers(req *DescribeLoadBalancersRequest) (*DescribeLoadBalancersResponse, error) {
	return c.DescribeLoadBalancersWithContext(context.Background(), req)
}

func (c *Client) DescribeLoadBalancersWithContext(ctx context.Context, req *DescribeLoadBalancersRequest) (*DescribeLoadBalancersResponse, error) {
	params := &struct {
		DescribeLoadBalancersRequest `json:",inline"`
		Action                       string
		Version                      string
	}{
		DescribeLoadBalancersRequest: *req,
		Action:                       "DescribeLoadBalancers",
		Version:                      "2015-12-01",
	}

	httpreq, err := c.newRequest(params)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &DescribeLoadBalancersResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
