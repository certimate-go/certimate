package nginxproxymanager

import (
	"context"
	"net/http"
)

type SettingsSetDefaultSiteRequest struct {
	Value string `json:"value"`
	Meta  struct {
		Redirect string `json:"redirect"`
		Html     string `json:"html"`
	} `json:"meta"`
}

type SettingsSetDefaultSiteResponse struct{}

func (c *Client) SettingsSetDefaultSite(req *SettingsSetDefaultSiteRequest) (*SettingsSetDefaultSiteResponse, error) {
	return c.SettingsSetDefaultSiteWithContext(context.Background(), req)
}

func (c *Client) SettingsSetDefaultSiteWithContext(ctx context.Context, req *SettingsSetDefaultSiteRequest) (*SettingsSetDefaultSiteResponse, error) {
	if err := c.ensureJwtTokenExists(); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodPut, "/settings/default-site")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &SettingsSetDefaultSiteResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
