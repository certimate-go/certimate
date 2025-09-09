package btwaf

import (
	"context"
	"net/http"
)

type ModifySiteRequest struct {
	SiteId *string            `json:"site_id,omitempty"`
	Type   *string            `json:"types,omitempty"`
	Server *SiteServerInfoMod `json:"server,omitempty"`
}

type ModifySiteResponse struct {
	apiResponseBase
}

func (c *Client) ModifySite(req *ModifySiteRequest) (*ModifySiteResponse, error) {
	return c.ModifySiteWithContext(context.Background(), req)
}

func (c *Client) ModifySiteWithContext(ctx context.Context, req *ModifySiteRequest) (*ModifySiteResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/wafmastersite/modify_site")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &ModifySiteResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
