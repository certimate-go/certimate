package nginxproxymanager

import (
	"context"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type NginxListCertificatesRequest struct {
	Expand *string `json:"expand,omitempty" url:"expand,omitempty"`
}

type NginxListCertificatesResponse = []*CertificateRecord

func (c *Client) NginxListCertificates(req *NginxListCertificatesRequest) (*NginxListCertificatesResponse, error) {
	return c.NginxListCertificatesWithContext(context.Background(), req)
}

func (c *Client) NginxListCertificatesWithContext(ctx context.Context, req *NginxListCertificatesRequest) (*NginxListCertificatesResponse, error) {
	if err := c.ensureJwtTokenExists(); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodGet, "/nginx/certificates")
	if err != nil {
		return nil, err
	} else {
		values, err := qs.Values(req)
		if err != nil {
			return nil, err
		}

		httpreq.SetQueryParamsFromValues(values)
		httpreq.SetContext(ctx)
	}

	result := &NginxListCertificatesResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
