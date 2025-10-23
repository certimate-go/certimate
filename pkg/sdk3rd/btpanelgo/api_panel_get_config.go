package btpanel

import (
	"context"
	"net/http"
)

type PanelGetConfigRequest struct{}

type PanelGetConfigResponse struct {
	apiResponseBase

	Paths *struct {
		Panel string `json:"panel,omitempty"`
		Soft  string `json:"soft,omitempty"`
	} `json:"paths,omitempty"`
	Site *struct {
		WebServer  string `json:"webserver,omitempty"`
		SitesPath  string `json:"sites_path,omitempty"`
		BackupPath string `json:"backup_path,omitempty"`
	} `json:"site,omitempty"`
}

func (c *Client) PanelGetConfig(req *PanelGetConfigRequest) (*PanelGetConfigResponse, error) {
	return c.PanelGetConfigWithContext(context.Background(), req)
}

func (c *Client) PanelGetConfigWithContext(ctx context.Context, req *PanelGetConfigRequest) (*PanelGetConfigResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/panel/get_config", req, false)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &PanelGetConfigResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
