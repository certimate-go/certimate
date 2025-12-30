package nginxproxymanager

import (
	"context"
	"fmt"
	"net/http"
)

type NginxUpdateRedirectionHostRequest struct {
	CertificateId *int64 `json:"certificate_id,omitempty"`
}

type NginxUpdateRedirectionHostResponse struct {
	RedirectionHostRecord
}

func (c *Client) NginxUpdateRedirectionHost(hostId int64, req *NginxUpdateRedirectionHostRequest) (*NginxUpdateRedirectionHostResponse, error) {
	return c.NginxUpdateRedirectionHostWithContext(context.Background(), hostId, req)
}

func (c *Client) NginxUpdateRedirectionHostWithContext(ctx context.Context, hostId int64, req *NginxUpdateRedirectionHostRequest) (*NginxUpdateRedirectionHostResponse, error) {
	if hostId == 0 {
		return nil, fmt.Errorf("sdkerr: unset hostId")
	}

	if err := c.ensureJwtTokenExists(); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/nginx/redirection-hosts/%d", hostId))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &NginxUpdateRedirectionHostResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
