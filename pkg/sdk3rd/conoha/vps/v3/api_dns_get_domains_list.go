package v3

import (
	"context"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type DnsGetDomainsListRequest struct {
	Limit    *int    `json:"limit,omitempty"     url:"limit,omitempty"`
	Offset   *int    `json:"offset,omitempty"    url:"offset,omitempty"`
	SortType *string `json:"sort_type,omitempty" url:"sort_type,omitempty"`
	SortKey  *string `json:"sort_key,omitempty"  url:"sort_key,omitempty"`
}

type DnsGetDomainsListResponse struct {
	sdkResponseBase

	Domains    []*Domain `json:"domains,omitempty"`
	TotalCount int       `json:"total_count,omitempty"`
}

func (c *Client) DnsGetDomainsList(req *DnsGetDomainsListRequest) (*DnsGetDomainsListResponse, error) {
	return c.DnsGetDomainsListWithContext(context.Background(), req)
}

func (c *Client) DnsGetDomainsListWithContext(ctx context.Context, req *DnsGetDomainsListRequest) (*DnsGetDomainsListResponse, error) {
	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	path := dnsBaseURL + "/v1/domains"
	httpreq, err := c.newRequest(http.MethodGet, path)
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

	result := &DnsGetDomainsListResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
