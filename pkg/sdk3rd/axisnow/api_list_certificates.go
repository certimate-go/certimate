package axisnow

import (
	"context"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type ListCertificatesRequest struct {
	Filter  *string `json:"filter,omitempty"   url:"filter,omitempty"`
	Page    *int    `json:"page,omitempty"     url:"page,omitempty"`
	PerPage *int    `json:"per_page,omitempty" url:"per_page,omitempty"`
}

type ListCertificatesResponse struct {
	sdkResponseBase

	ResultInfo *struct {
		Page            int `json:"page"`
		PerPage         int `json:"per_page"`
		Count           int `json:"count"`
		TotalCount      int `json:"total_count"`
		SubscribedTotal int `json:"subscribed_total"`
	} `json:"result_info,omitempty"`
	Results []*Certificate `json:"result,omitempty"`
}

func (c *Client) ListCertificates(req *ListCertificatesRequest) (*ListCertificatesResponse, error) {
	return c.ListCertificatesWithContext(context.Background(), req)
}

func (c *Client) ListCertificatesWithContext(ctx context.Context, req *ListCertificatesRequest) (*ListCertificatesResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, "/certificates")
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
