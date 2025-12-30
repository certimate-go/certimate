package nginxproxymanager

import (
	"context"
	"fmt"
	"net/http"
)

type NginxUpdateStreamRequest struct {
	CertificateId *int64 `json:"certificate_id,omitempty"`
}

type NginxUpdateStreamResponse struct {
	StreamHostRecord
}

func (c *Client) NginxUpdateStream(hostId int64, req *NginxUpdateStreamRequest) (*NginxUpdateStreamResponse, error) {
	return c.NginxUpdateStreamWithContext(context.Background(), hostId, req)
}

func (c *Client) NginxUpdateStreamWithContext(ctx context.Context, hostId int64, req *NginxUpdateStreamRequest) (*NginxUpdateStreamResponse, error) {
	if hostId == 0 {
		return nil, fmt.Errorf("sdkerr: unset hostId")
	}

	if err := c.ensureJwtTokenExists(); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/nginx/streams/%d", hostId))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &NginxUpdateStreamResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
