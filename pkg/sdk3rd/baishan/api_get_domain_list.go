package baishan

import (
	"context"
	"encoding/json"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type GetDomainListRequest struct {
	PageNumber   *int32  `json:"page_number,omitempty"   url:"page_number,omitempty"`
	PageSize     *int32  `json:"page_size,omitempty"     url:"page_size,omitempty"`
	DomainStatus *string `json:"domain_status,omitempty" url:"domain_status,omitempty"`
}

type GetDomainListResponse struct {
	sdkResponseBase

	Data []*struct {
		List        []*Domain   `json:"list"`
		PageNumber  json.Number `json:"page_number"`
		PageSize    json.Number `json:"page_size"`
		TotalNumber json.Number `json:"total_number"`
	} `json:"data,omitempty"`
}

func (c *Client) GetDomainList(req *GetDomainListRequest) (*GetDomainListResponse, error) {
	return c.GetDomainListWithContext(context.Background(), req)
}

func (c *Client) GetDomainListWithContext(ctx context.Context, req *GetDomainListRequest) (*GetDomainListResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, "/v2/domain/list")
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

	result := &GetDomainListResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
