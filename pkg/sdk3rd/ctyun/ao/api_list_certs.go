package ao

import (
	"context"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type ListCertsRequest struct {
	Page      *int32 `json:"page,omitempty"       url:"page,omitempty"`
	PerPage   *int32 `json:"per_page,omitempty"   url:"per_page,omitempty"`
	UsageMode *int32 `json:"usage_mode,omitempty" url:"usage_mode,omitempty"`
}

type ListCertsResponse struct {
	sdkResponseBase

	ReturnObj *struct {
		Results      []*Cert `json:"result,omitempty"`
		Page         int32   `json:"page,omitempty"`
		PerPage      int32   `json:"per_page,omitempty"`
		TotalPage    int32   `json:"total_page,omitempty"`
		TotalRecords int32   `json:"total_records,omitempty"`
	} `json:"returnObj,omitempty"`
}

func (c *Client) ListCerts(req *ListCertsRequest) (*ListCertsResponse, error) {
	return c.ListCertsWithContext(context.Background(), req)
}

func (c *Client) ListCertsWithContext(ctx context.Context, req *ListCertsRequest) (*ListCertsResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, "/ctapi/v1/accessone/cert/list")
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

	result := &ListCertsResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
