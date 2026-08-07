package icdn

import (
	"context"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type QueryDomainListRequest struct {
	Page        *int32  `json:"page,omitempty"         url:"page,omitempty"`
	PageSize    *int32  `json:"page_size,omitempty"    url:"page_size,omitempty"`
	Domain      *string `json:"domain,omitempty"       url:"domain,omitempty"`
	ProductCode *string `json:"product_code,omitempty" url:"product_code,omitempty"`
	Status      *int32  `json:"status,omitempty"       url:"status,omitempty"`
	AreaScope   *int32  `json:"area_scope,omitempty"   url:"area_scope,omitempty"`
}

type QueryDomainListResponse struct {
	sdkResponseBase

	ReturnObj *struct {
		Results   []*Domain `json:"result,omitempty"`
		Page      int32     `json:"page,omitempty"`
		PageSize  int32     `json:"page_size,omitempty"`
		PageCount int32     `json:"page_count,omitempty"`
		Total     int32     `json:"total,omitempty"`
	} `json:"returnObj,omitempty"`
}

func (c *Client) QueryDomainList(req *QueryDomainListRequest) (*QueryDomainListResponse, error) {
	return c.QueryDomainListWithContext(context.Background(), req)
}

func (c *Client) QueryDomainListWithContext(ctx context.Context, req *QueryDomainListRequest) (*QueryDomainListResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, "/v1/domain/query-domain-list")
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

	result := &QueryDomainListResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
