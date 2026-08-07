package certificate

import (
	"context"
	"net/http"
)

type ListCertificatesResponse struct {
	sdkResponseBase

	Certificates []*Certificate `json:"ssl-certificates,omitempty"`
}

func (c *Client) ListCertificates() (*ListCertificatesResponse, error) {
	return c.ListCertificatesWithContext(context.Background())
}

func (c *Client) ListCertificatesWithContext(ctx context.Context) (*ListCertificatesResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, "/api/ssl/certificate")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &ListCertificatesResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
