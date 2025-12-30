package nginxproxymanager

import (
	"context"
	"fmt"
	"net/http"
)

type NginxUpdateDeadHostRequest struct {
	CertificateId *int64 `json:"certificate_id,omitempty"`
}

type NginxUpdateDeadHostResponse struct {
	DeadHostRecord
}

func (c *Client) NginxUpdateDeadHost(hostId int64, req *NginxUpdateDeadHostRequest) (*NginxUpdateDeadHostResponse, error) {
	return c.NginxUpdateDeadHostWithContext(context.Background(), hostId, req)
}

func (c *Client) NginxUpdateDeadHostWithContext(ctx context.Context, hostId int64, req *NginxUpdateDeadHostRequest) (*NginxUpdateDeadHostResponse, error) {
	if hostId == 0 {
		return nil, fmt.Errorf("sdkerr: unset hostId")
	}

	if err := c.ensureJwtTokenExists(); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/nginx/dead-hosts/%d", hostId))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &NginxUpdateDeadHostResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
