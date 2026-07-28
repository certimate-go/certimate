package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type ListServerCertificatesRequest = iam.ListServerCertificatesInput

type ListServerCertificatesResponse = iam.ListServerCertificatesOutput

func (c *Client) ListServerCertificates(req *ListServerCertificatesRequest) (*ListServerCertificatesResponse, error) {
	return c.ListServerCertificatesWithContext(context.Background(), req)
}

func (c *Client) ListServerCertificatesWithContext(ctx context.Context, req *ListServerCertificatesRequest) (*ListServerCertificatesResponse, error) {
	params := &struct {
		ListServerCertificatesRequest `json:",inline"`
		Action                        string
		Version                       string
	}{
		ListServerCertificatesRequest: *req,
		Action:                        "ListServerCertificates",
		Version:                       "2010-05-08",
	}

	httpreq, err := c.newRequest(params)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &ListServerCertificatesResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
