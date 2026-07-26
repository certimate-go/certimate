package lvdn

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

	Results   []*Domain `json:"result,omitempty"`
	Page      int32     `json:"page,omitempty"`
	PageSize  int32     `json:"page_size,omitempty"`
	PageCount int32     `json:"page_count,omitempty"`
	Total     int32     `json:"total,omitempty"`

	ReturnObj *QueryDomainListResult `json:"returnObj,omitempty"`
}

type QueryDomainListResult struct {
	Results   []*Domain `json:"result,omitempty"`
	Page      int32     `json:"page,omitempty"`
	PageSize  int32     `json:"page_size,omitempty"`
	PageCount int32     `json:"page_count,omitempty"`
	Total     int32     `json:"total,omitempty"`
}

func (c *Client) QueryDomainList(req *QueryDomainListRequest) (*QueryDomainListResponse, error) {
	return c.QueryDomainListWithContext(context.Background(), req)
}

func (c *Client) QueryDomainListWithContext(ctx context.Context, req *QueryDomainListRequest) (*QueryDomainListResponse, error) {
	httpreq, err := c.domainClient.NewRequest(http.MethodGet, "/domain/query-domain-list")
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
	if _, err := c.doDomainRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	result.ReturnObj = &QueryDomainListResult{
		Results:   result.Results,
		Page:      result.Page,
		PageSize:  result.PageSize,
		PageCount: result.PageCount,
		Total:     result.Total,
	}

	return result, nil
}
