package btpanel

import (
	"context"
	"net/http"
)

type DatalistGetDataListRequest struct {
	Table        *string `json:"table,omitempty"`
	SearchType   *string `json:"search_type,omitempty"`
	SearchString *string `json:"search,omitempty"`
	Page         *int32  `json:"p,omitempty"`
	Limit        *int32  `json:"limit,omitempty"`
	Order        *string `json:"order,omitempty"`
	Type         *int32  `json:"type,omitempty"`
}

type DatalistGetDataListResponse struct {
	apiResponseBase
	Data []*SiteData `json:"data,omitempty"`
	Page *PageData   `json:"page,omitempty"`
}

func (c *Client) DatalistGetDataList(req *DatalistGetDataListRequest) (*DatalistGetDataListResponse, error) {
	return c.DatalistGetDataListWithContext(context.Background(), req)
}

func (c *Client) DatalistGetDataListWithContext(ctx context.Context, req *DatalistGetDataListRequest) (*DatalistGetDataListResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/datalist/get_data_list", req, false)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &DatalistGetDataListResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
