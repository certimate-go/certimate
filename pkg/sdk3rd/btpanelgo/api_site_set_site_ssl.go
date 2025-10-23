package btpanel

import (
	"context"
	"net/http"
)

type SiteSetSiteSSLRequest struct {
	SiteId *int32  `json:"siteid,omitempty"`
	Status *bool   `json:"status,omitempty"`
	Key    *string `json:"key,omitempty"`
	Cert   *string `json:"cert,omitempty"`
}

type SiteSetSiteSSLResponse struct {
	apiResponseBase
}

func (c *Client) SiteSetSiteSSL(req *SiteSetSiteSSLRequest) (*SiteSetSiteSSLResponse, error) {
	return c.SiteSetSiteSSLWithContext(context.Background(), req)
}

func (c *Client) SiteSetSiteSSLWithContext(ctx context.Context, req *SiteSetSiteSSLRequest) (*SiteSetSiteSSLResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/site/set_site_ssl", req, false)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &SiteSetSiteSSLResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
