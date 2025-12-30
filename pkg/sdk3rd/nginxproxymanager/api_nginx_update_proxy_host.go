package nginxproxymanager

import (
	"context"
	"fmt"
	"net/http"
)

type NginxUpdateProxyHostRequest struct {
	CertificateId *int64 `json:"certificate_id,omitempty"`
}

type NginxUpdateProxyHostResponse struct {
	ProxyHostRecord
}

func (c *Client) NginxUpdateProxyHost(hostId int64, req *NginxUpdateProxyHostRequest) (*NginxUpdateProxyHostResponse, error) {
	return c.NginxUpdateProxyHostWithContext(context.Background(), hostId, req)
}

func (c *Client) NginxUpdateProxyHostWithContext(ctx context.Context, hostId int64, req *NginxUpdateProxyHostRequest) (*NginxUpdateProxyHostResponse, error) {
	if hostId == 0 {
		return nil, fmt.Errorf("sdkerr: unset hostId")
	}

	if err := c.ensureJwtTokenExists(); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/nginx/proxy-hosts/%d", hostId))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &NginxUpdateProxyHostResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
