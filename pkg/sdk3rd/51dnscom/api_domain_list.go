package dnscom

import (
	"context"
	"net/http"
)

type DomainListRequest struct {
	GroupID  *int64 `json:"groupID,omitempty"`
	Page     *int32 `json:"page,omitempty"`
	PageSize *int32 `json:"pageSize,omitempty"`
}

type DomainListResponse struct {
	apiResponseBase

	Data *struct {
		Data        []*DomainRecord `json:"data"`
		RecordCount int32           `json:"recordCount"`
		Page        int32           `json:"page"`
		PageSize    int32           `json:"pageSize"`
		PageCount   int32           `json:"pageCount"`
	} `json:"data"`
}

func (c *Client) DomainList(req *DomainListRequest) (*DomainListResponse, error) {
	return c.DomainListWithContext(context.Background(), req)
}

func (c *Client) DomainListWithContext(ctx context.Context, req *DomainListRequest) (*DomainListResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/domain/list/", req)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &DomainListResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
