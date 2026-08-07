package teomakers

import (
	"context"
)

type DescribePagesZoneCustomDomainsRequest struct {
	ProjectId *string `json:"ProjectId,omitempty"`
}

type DescribePagesZoneCustomDomainsResponse struct {
	sdkResponseBase

	Data *struct {
		Response *struct {
			RequestId    *string                  `json:"RequestId,omitempty"`
			TotalCount   *int64                   `json:"TotalCount,omitempty"`
			PagesDomains []*PagesZoneCustomDomain `json:"PagesDomains,omitempty"`
		} `json:"Response,omitempty"`
	} `json:"Data,omitempty"`
}

func (c *Client) DescribePagesZoneCustomDomains(req *DescribePagesZoneCustomDomainsRequest) (*DescribePagesZoneCustomDomainsResponse, error) {
	return c.DescribePagesZoneCustomDomainsWithContext(context.Background(), req)
}

func (c *Client) DescribePagesZoneCustomDomainsWithContext(ctx context.Context, req *DescribePagesZoneCustomDomainsRequest) (*DescribePagesZoneCustomDomainsResponse, error) {
	httpreq, err := c.newRequest(req, "DescribePagesZoneCustomDomains")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &DescribePagesZoneCustomDomainsResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
