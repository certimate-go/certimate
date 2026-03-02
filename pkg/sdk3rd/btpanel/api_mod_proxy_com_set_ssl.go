package btpanel

import (
	"context"
	"net/http"
)

type ModProxyComSetSSLRequest struct {
	SiteName    string `json:"site_name"`
	PrivateKey  string `json:"key"`
	Certificate string `json:"csr"`
}

type ModProxyComSetSSLResponse struct {
	sdkResponseBase
}

func (c *Client) ModProxyComSetSSL(req *ModProxyComSetSSLRequest) (*ModProxyComSetSSLResponse, error) {
	return c.ModProxyComSetSSLWithContext(context.Background(), req)
}

func (c *Client) ModProxyComSetSSLWithContext(ctx context.Context, req *ModProxyComSetSSLRequest) (*ModProxyComSetSSLResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/mod/proxy/com/set_ssl", req)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &ModProxyComSetSSLResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
