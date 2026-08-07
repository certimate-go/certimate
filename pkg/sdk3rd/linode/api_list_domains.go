package linode

import (
	"context"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type ListDomainsRequest struct {
	Page     *int `json:"page,omitempty"      url:"page,omitempty"`
	PageSize *int `json:"page_size,omitempty" url:"page_size,omitempty"`
}

type ListDomainsResponse struct {
	sdkResponseBase

	Data    []*Domain `json:"data,omitempty"`
	Page    int       `json:"page,omitempty"`
	Pages   int       `json:"pages,omitempty"`
	Results int       `json:"results,omitempty"`
}

func (c *Client) ListDomains(req *ListDomainsRequest) (*ListDomainsResponse, error) {
	return c.ListDomainsWithContext(context.Background(), req)
}

func (c *Client) ListDomainsWithContext(ctx context.Context, req *ListDomainsRequest) (*ListDomainsResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, "/domains")
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

	result := &ListDomainsResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
