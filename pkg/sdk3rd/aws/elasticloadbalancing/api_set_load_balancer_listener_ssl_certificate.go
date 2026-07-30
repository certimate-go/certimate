package elasticloadbalancing

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
)

type SetLoadBalancerListenerSSLCertificateRequest = elasticloadbalancing.SetLoadBalancerListenerSSLCertificateInput

type SetLoadBalancerListenerSSLCertificateResponse = elasticloadbalancing.SetLoadBalancerListenerSSLCertificateOutput

func (c *Client) SetLoadBalancerListenerSSLCertificate(req *SetLoadBalancerListenerSSLCertificateRequest) (*SetLoadBalancerListenerSSLCertificateResponse, error) {
	return c.SetLoadBalancerListenerSSLCertificateWithContext(context.Background(), req)
}

func (c *Client) SetLoadBalancerListenerSSLCertificateWithContext(ctx context.Context, req *SetLoadBalancerListenerSSLCertificateRequest) (*SetLoadBalancerListenerSSLCertificateResponse, error) {
	params := &struct {
		SetLoadBalancerListenerSSLCertificateRequest `json:",inline"`
		Action                                       string
		Version                                      string
	}{
		SetLoadBalancerListenerSSLCertificateRequest: *req,
		Action:  "SetLoadBalancerListenerSSLCertificate",
		Version: "2012-06-01",
	}

	httpreq, err := c.newRequest(params)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &SetLoadBalancerListenerSSLCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
