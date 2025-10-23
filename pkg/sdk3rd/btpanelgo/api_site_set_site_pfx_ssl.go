package btpanel

import (
	"context"
	"net/http"
)

type SiteSetSitePFXSSLRequest struct {
	SiteId   *int32  `json:"siteid,omitempty"`
	PFX      *string `json:"pfx,omitempty"`
	Password *string `json:"password,omitempty"`
}

type SiteSetSitePFXSSLResponse struct {
	apiResponseBase
}

func (c *Client) SiteSetSitePFXSSL(req *SiteSetSitePFXSSLRequest) (*SiteSetSitePFXSSLResponse, error) {
	return c.SiteSetSitePFXSSLWithContext(context.Background(), req)
}

func (c *Client) SiteSetSitePFXSSLWithContext(ctx context.Context, req *SiteSetSitePFXSSLRequest) (*SiteSetSitePFXSSLResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/site/set_site_pfx_ssl", req, false)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &SiteSetSitePFXSSLResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
