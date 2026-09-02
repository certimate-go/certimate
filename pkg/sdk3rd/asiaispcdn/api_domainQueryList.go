package asiaispcdn

import (
	"context"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type DomainQueryListRequest struct {
	Domain     *string `json:"domain,omitempty"     url:"domain,omitempty"`
	SubDomain  *bool   `json:"subDomain,omitempty"  url:"subDomain,omitempty"`
	LiveDomain *bool   `json:"liveDomain,omitempty" url:"liveDomain,omitempty"`
}

type DomainQueryListResponse struct {
	sdkResponseBase

	Data []*Domain `json:"data,omitempty"`
}

func (c *Client) DomainQueryList(req *DomainQueryListRequest) (*DomainQueryListResponse, error) {
	return c.DomainQueryListWithContext(context.Background(), req)
}

func (c *Client) DomainQueryListWithContext(ctx context.Context, req *DomainQueryListRequest) (*DomainQueryListResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, nil, "domainQueryList")
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

	result := &DomainQueryListResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
