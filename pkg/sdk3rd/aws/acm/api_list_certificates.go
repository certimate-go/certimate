package acm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/acm"
)

type ListCertificatesRequest = acm.ListCertificatesInput

type ListCertificatesResponse = acm.ListCertificatesOutput

func (c *Client) ListCertificates(req *ListCertificatesRequest) (*ListCertificatesResponse, error) {
	return c.ListCertificatesWithContext(context.Background(), req)
}

func (c *Client) ListCertificatesWithContext(ctx context.Context, req *ListCertificatesRequest) (*ListCertificatesResponse, error) {
	httpreq, err := c.newRequest(buildAmzTarget("ListCertificates"))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &ListCertificatesResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
