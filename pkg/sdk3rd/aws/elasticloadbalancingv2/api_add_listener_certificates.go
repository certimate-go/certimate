package elasticloadbalancingv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
)

type AddListenerCertificatesRequest = elasticloadbalancingv2.AddListenerCertificatesInput

type AddListenerCertificatesResponse = elasticloadbalancingv2.AddListenerCertificatesOutput

func (c *Client) AddListenerCertificates(req *AddListenerCertificatesRequest) (*AddListenerCertificatesResponse, error) {
	return c.AddListenerCertificatesWithContext(context.Background(), req)
}

func (c *Client) AddListenerCertificatesWithContext(ctx context.Context, req *AddListenerCertificatesRequest) (*AddListenerCertificatesResponse, error) {
	params := &struct {
		AddListenerCertificatesRequest `json:",inline"`
		Action                         string
		Version                        string
	}{
		AddListenerCertificatesRequest: *req,
		Action:                         "AddListenerCertificates",
		Version:                        "2015-12-01",
	}

	httpreq, err := c.newRequest(params)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &AddListenerCertificatesResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
