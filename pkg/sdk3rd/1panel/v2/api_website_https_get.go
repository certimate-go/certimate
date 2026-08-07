package v2

import (
	"context"
	"fmt"
	"net/http"
)

type WebsiteHttpsGetResponse struct {
	sdkResponseBase

	Data *WebsiteHTTPSConfig `json:"data,omitempty"`
}

func (c *Client) WebsiteHttpsGet(websiteId int64) (*WebsiteHttpsGetResponse, error) {
	return c.WebsiteHttpsGetWithContext(context.Background(), websiteId)
}

func (c *Client) WebsiteHttpsGetWithContext(ctx context.Context, websiteId int64) (*WebsiteHttpsGetResponse, error) {
	if websiteId == 0 {
		return nil, fmt.Errorf("sdkerr: bad request: unset websiteId")
	}

	httpreq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/websites/%d/https", websiteId))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &WebsiteHttpsGetResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
