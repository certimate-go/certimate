package btpanelgo

import (
	"context"
	"net/http"
)

type SiteGetProjectListRequest struct {
	SearchType   *string `json:"search_type,omitempty"`
	SearchString *string `json:"search,omitempty"`
	Page         *int32  `json:"p,omitempty"`
	Limit        *int32  `json:"limit,omitempty"`
	Order        *string `json:"order,omitempty"`
}

type SiteGetProjectListResponse struct {
	sdkResponseBase
	Data []*SiteData `json:"data,omitempty"`
	Page *PageData   `json:"page,omitempty"`
}

func (c *Client) SiteGetProjectList(req *SiteGetProjectListRequest) (*SiteGetProjectListResponse, error) {
	return c.SiteGetProjectListWithContext(context.Background(), req)
}

func (c *Client) SiteGetProjectListWithContext(ctx context.Context, req *SiteGetProjectListRequest) (*SiteGetProjectListResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/site/get_project_list", req, false)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &SiteGetProjectListResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
