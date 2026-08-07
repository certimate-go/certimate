package cdnfly

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type GetSiteResponse struct {
	sdkResponseBase

	Data *struct {
		Id          int64  `json:"id"`
		Name        string `json:"name"`
		Domain      string `json:"domain"`
		HttpsListen string `json:"https_listen"`
	} `json:"data,omitempty"`
}

func (c *Client) GetSite(siteId string) (*GetSiteResponse, error) {
	return c.GetSiteWithContext(context.Background(), siteId)
}

func (c *Client) GetSiteWithContext(ctx context.Context, siteId string) (*GetSiteResponse, error) {
	if siteId == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset siteId")
	}

	httpreq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/sites/%s", url.PathEscape(siteId)))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &GetSiteResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
