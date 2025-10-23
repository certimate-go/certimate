package btpanel

import (
	"context"
	"net/http"
)

type ConfigSetPanelSSLRequest struct {
	SSLStatus *int32  `json:"ssl_status,omitempty"`
	SSLKey    *string `json:"ssl_key,omitempty"`
	SSLPem    *string `json:"ssl_pem,omitempty"`
}

type ConfigSetPanelSSLResponse struct {
	apiResponseBase
}

func (c *Client) ConfigSetPanelSSL(req *ConfigSetPanelSSLRequest) (*ConfigSetPanelSSLResponse, error) {
	return c.ConfigSetPanelSSLWithContext(context.Background(), req)
}

func (c *Client) ConfigSetPanelSSLWithContext(ctx context.Context, req *ConfigSetPanelSSLRequest) (*ConfigSetPanelSSLResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/config/set_panel_ssl", req, false)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &ConfigSetPanelSSLResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
