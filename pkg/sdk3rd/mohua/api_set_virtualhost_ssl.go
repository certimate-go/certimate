package mohua

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type SetVirtualHostSSLRequest struct {
	ID       int    `json:"id"`
	SSLForce string `json:"ssl_force"`
	SSLCert  string `json:"sslCert"`
	SSLKey   string `json:"sslKey"`
}

type SetVirtualHostSSLResponse struct {
	sdkResponseBase

	Data []*DomainInfo `json:"data"`
}

func (c *Client) SetVirtualHostSSL(hostId string, req *SetVirtualHostSSLRequest) (*SetVirtualHostSSLResponse, error) {
	return c.SetVirtualHostSSLWithContext(context.Background(), hostId, req)
}

func (c *Client) SetVirtualHostSSLWithContext(ctx context.Context, hostId string, req *SetVirtualHostSSLRequest) (*SetVirtualHostSSLResponse, error) {
	if hostId == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset hostId")
	}

	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/provision/custom/%s/domains", url.PathEscape(hostId))
	httpreq, err := c.newRequest(http.MethodPost, path)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(map[string]any{
			"func":      "SetSSL",
			"id":        req.ID,
			"ssl_force": req.SSLForce,
			"sslCert":   url.QueryEscape(req.SSLCert),
			"sslKey":    url.QueryEscape(req.SSLKey),
		})
		httpreq.SetContext(ctx)
	}

	result := &SetVirtualHostSSLResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
