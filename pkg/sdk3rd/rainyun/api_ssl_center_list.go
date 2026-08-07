package rainyun

import (
	"context"
	"encoding/json"
	"net/http"
)

type SslCenterListFilters struct {
	Domain *string `json:"Domain,omitempty"`
}

type SslCenterListRequest struct {
	Filters *SslCenterListFilters `json:"columnFilters,omitempty"`
	Sort    []*string             `json:"sort,omitempty"`
	Page    *int32                `json:"page,omitempty"`
	PerPage *int32                `json:"perPage,omitempty"`
}

type SslCenterListResponse struct {
	sdkResponseBase

	Data *struct {
		TotalRecords int32             `json:"TotalRecords"`
		Records      []*SSLCertificate `json:"Records"`
	} `json:"data,omitempty"`
}

func (c *Client) SslCenterList(req *SslCenterListRequest) (*SslCenterListResponse, error) {
	return c.SslCenterListWithContext(context.Background(), req)
}

func (c *Client) SslCenterListWithContext(ctx context.Context, req *SslCenterListRequest) (*SslCenterListResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, "/product/sslcenter")
	if err != nil {
		return nil, err
	} else {
		jsonb, _ := json.Marshal(req)
		httpreq.SetQueryParam("options", string(jsonb))
		httpreq.SetContext(ctx)
	}

	result := &SslCenterListResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
