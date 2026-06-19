package onepanel

import (
	"context"
	"net/http"
)

type WebsiteSearchRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Order    string `json:"order"`
	OrderBy  string `json:"orderBy"`
	Page     int32  `json:"page"`
	PageSize int32  `json:"pageSize"`
}

type WebsiteSearchResponse struct {
	sdkResponseBase

	Data *struct {
		Items []*Website `json:"items"`
		Total int32      `json:"total"`
	} `json:"data,omitempty"`
}

func (c *Client) WebsiteSearch(req *WebsiteSearchRequest) (*WebsiteSearchResponse, error) {
	return c.WebsiteSearchWithContext(context.Background(), req)
}

func (c *Client) WebsiteSearchWithContext(ctx context.Context, req *WebsiteSearchRequest) (*WebsiteSearchResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/websites/search")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &WebsiteSearchResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
