package dns

import (
	"context"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type QueryRecordListRequest struct {
	Domain   *string `json:"domain,omitempty"   url:"domain,omitempty"`
	Host     *string `json:"host,omitempty"     url:"host,omitempty"`
	Type     *string `json:"type,omitempty"     url:"type,omitempty"`
	LineCode *string `json:"lineCode,omitempty" url:"lineCode,omitempty"`
	Value    *string `json:"value,omitempty"    url:"value,omitempty"`
	State    *int32  `json:"state,omitempty"    url:"state,omitempty"`
}

type QueryRecordListResponse struct {
	sdkResponseBase

	ReturnObj *struct {
		Records []*DNSRecord `json:"records,omitempty"`
	} `json:"returnObj,omitempty"`
}

func (c *Client) QueryRecordList(req *QueryRecordListRequest) (*QueryRecordListResponse, error) {
	return c.QueryRecordListWithContext(context.Background(), req)
}

func (c *Client) QueryRecordListWithContext(ctx context.Context, req *QueryRecordListRequest) (*QueryRecordListResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, "/v2/queryRecordList")
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

	result := &QueryRecordListResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
