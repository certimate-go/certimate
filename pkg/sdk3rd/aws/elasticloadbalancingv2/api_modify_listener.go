package elasticloadbalancingv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
)

type ModifyListenerRequest = elasticloadbalancingv2.ModifyListenerInput

type ModifyListenerResponse = elasticloadbalancingv2.ModifyListenerOutput

func (c *Client) ModifyListener(req *ModifyListenerRequest) (*ModifyListenerResponse, error) {
	return c.ModifyListenerWithContext(context.Background(), req)
}

func (c *Client) ModifyListenerWithContext(ctx context.Context, req *ModifyListenerRequest) (*ModifyListenerResponse, error) {
	params := &struct {
		ModifyListenerRequest `json:",inline"`
		Action                string
		Version               string
	}{
		ModifyListenerRequest: *req,
		Action:                "ModifyListener",
		Version:               "2015-12-01",
	}

	httpreq, err := c.newRequest(params)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &ModifyListenerResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
