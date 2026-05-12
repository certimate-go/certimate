package cdnfly

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type UpdateSiteRequest struct {
	HttpsListen *map[string]any `json:"https_listen,omitempty"`
	Enable      *bool           `json:"enable,omitempty"`
}

type UpdateSiteResponse struct {
	sdkResponseBase
}

func (c *Client) UpdateSite(siteId string, req *UpdateSiteRequest) (*UpdateSiteResponse, error) {
	return c.UpdateSiteWithContext(context.Background(), siteId, req)
}

func (c *Client) UpdateSiteWithContext(ctx context.Context, siteId string, req *UpdateSiteRequest) (*UpdateSiteResponse, error) {
	if siteId == "" {
		return nil, fmt.Errorf("sdkerr: unset siteId")
	}

	httpreq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/sites/%s", url.PathEscape(siteId)))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &UpdateSiteResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
