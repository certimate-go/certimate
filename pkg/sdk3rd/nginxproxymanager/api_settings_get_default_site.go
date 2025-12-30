package nginxproxymanager

import (
	"context"
	"net/http"
)

type SettingsGetDefaultSiteRequest struct{}

type SettingsGetDefaultSiteResponse struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Value       string `json:"value"`
	Meta        struct {
		Redirect string `json:"redirect"`
		Html     string `json:"urhtmll"`
	} `json:"meta"`
}

func (c *Client) SettingsGetDefaultSite(req *SettingsGetDefaultSiteRequest) (*SettingsGetDefaultSiteResponse, error) {
	return c.SettingsGetDefaultSiteWithContext(context.Background(), req)
}

func (c *Client) SettingsGetDefaultSiteWithContext(ctx context.Context, req *SettingsGetDefaultSiteRequest) (*SettingsGetDefaultSiteResponse, error) {
	if err := c.ensureJwtTokenExists(); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodGet, "/settings/default-site")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &SettingsGetDefaultSiteResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
