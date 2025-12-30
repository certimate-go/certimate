package dokploy

import (
	"context"
	"net/http"
)

type CertificatesAllRequest struct{}

type CertificatesAllResponse = []*Certificate

func (c *Client) CertificatesAll(req *CertificatesAllRequest) (*CertificatesAllResponse, error) {
	return c.CertificatesAllWithContext(context.Background(), req)
}

func (c *Client) CertificatesAllWithContext(ctx context.Context, req *CertificatesAllRequest) (*CertificatesAllResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, "/certificates.all")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &CertificatesAllResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
