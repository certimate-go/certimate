package elb

import (
	"context"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type ListCertificatesRequest struct {
	ClientToken *string `json:"clientToken,omitempty" url:"clientToken,omitempty"`
	RegionID    *string `json:"regionID,omitempty"    url:"regionID,omitempty"`
	IDs         *string `json:"IDs,omitempty"         url:"IDs,omitempty"`
	Name        *string `json:"name,omitempty"        url:"name,omitempty"`
	Type        *string `json:"type,omitempty"        url:"type,omitempty"`
}

type ListCertificatesResponse struct {
	sdkResponseBase

	ReturnObj []*Certificate `json:"returnObj,omitempty"`
}

func (c *Client) ListCertificates(req *ListCertificatesRequest) (*ListCertificatesResponse, error) {
	return c.ListCertificatesWithContext(context.Background(), req)
}

func (c *Client) ListCertificatesWithContext(ctx context.Context, req *ListCertificatesRequest) (*ListCertificatesResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, "/v4/elb/list-certificate")
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

	result := &ListCertificatesResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
