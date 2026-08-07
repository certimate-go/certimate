package v2

import (
	"context"
	"fmt"
	"net/http"
)

type WebsiteGetRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Page     int32  `json:"page"`
	PageSize int32  `json:"pageSize"`
}

type WebsiteGetResponse struct {
	sdkResponseBase

	Data *WebsiteDetail `json:"data,omitempty"`
}

func (c *Client) WebsiteGet(websiteId int64) (*WebsiteGetResponse, error) {
	return c.WebsiteGetWithContext(context.Background(), websiteId)
}

func (c *Client) WebsiteGetWithContext(ctx context.Context, websiteId int64) (*WebsiteGetResponse, error) {
	if websiteId == 0 {
		return nil, fmt.Errorf("sdkerr: bad request: unset websiteId")
	}

	httpreq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/websites/%d", websiteId))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &WebsiteGetResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
