package kcm

import (
	"context"
	"net/http"
)

type DescribeCertificatesRequest struct {
	Region   *string `json:"Region,omitempty"`
	Page     *int32  `json:"Page,omitempty"`
	PageSize *int32  `json:"PageSize,omitempty"`
}

type DescribeCertificatesResponse struct {
	sdkResponseBase

	CertificateSet []*LBCertificate `json:"CertificateSet"`
}

func (c *Client) DescribeCertificates(req *DescribeCertificatesRequest) (*DescribeCertificatesResponse, error) {
	return c.DescribeCertificatesWithContext(context.Background(), req)
}

func (c *Client) DescribeCertificatesWithContext(ctx context.Context, req *DescribeCertificatesRequest) (*DescribeCertificatesResponse, error) {
	params := &struct {
		DescribeCertificatesRequest `json:",inline"`
		Action                      string
		Version                     string
	}{
		DescribeCertificatesRequest: *req,
		Action:                      "DescribeCertificates",
		Version:                     "2016-03-04",
	}

	httpreq, err := c.newRequest(http.MethodGet, "/", params)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &DescribeCertificatesResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
