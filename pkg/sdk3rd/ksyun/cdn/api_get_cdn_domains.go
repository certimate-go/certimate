package cdn

import (
	"context"
	"net/http"
)

type GetCDNDomainsRequest struct {
	ProjectId    *int64  `json:"ProjectId,omitempty"`
	DomainName   *string `json:"DomainName,omitempty"`
	DomainStatus *string `json:"DomainStatus,omitempty"`
	FuzzyMatch   *bool   `json:"FuzzyMatch,omitempty"`
	CdnType      *bool   `json:"CdnType,omitempty"`
	PageNumber   *int32  `json:"PageNumber,omitempty"`
	PageSize     *int32  `json:"PageSize,omitempty"`
}

type GetCdnDomainsResponse struct {
	sdkResponseBase

	Domains    []*CDNDomain `json:"Domains"`
	PageNumber int32        `json:"PageNumber,omitempty"`
	PageSize   int32        `json:"PageSize,omitempty"`
	TotalCount int32        `json:"TotalCount,omitempty"`
}

func (c *Client) GetCDNDomains(req *GetCDNDomainsRequest) (*GetCdnDomainsResponse, error) {
	return c.GetCDNDomainsWithContext(context.Background(), req)
}

func (c *Client) GetCDNDomainsWithContext(ctx context.Context, req *GetCDNDomainsRequest) (*GetCdnDomainsResponse, error) {
	params := &struct {
		GetCDNDomainsRequest `json:",inline"`
		Action               string
		Version              string
	}{
		GetCDNDomainsRequest: *req,
		Action:               "GetCdnDomains",
		Version:              "2019-06-01",
	}

	httpreq, err := c.newRequest(http.MethodGet, "/2019-06-01/GetCdnDomains", params)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &GetCdnDomainsResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
