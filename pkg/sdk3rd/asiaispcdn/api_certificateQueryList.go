package asiaispcdn

import (
	"context"
	"net/http"
)

type CertificateQueryListResponse struct {
	sdkResponseBase

	Data []*Certificate `json:"data,omitempty"`
}

func (c *Client) CertificateQueryList() (*CertificateQueryListResponse, error) {
	return c.CertificateQueryListWithContext(context.Background())
}

func (c *Client) CertificateQueryListWithContext(ctx context.Context) (*CertificateQueryListResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, nil, "certificateQueryList")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &CertificateQueryListResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
