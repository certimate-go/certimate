package rainyun

import (
	"context"
	"fmt"
	"net/http"
)

type RcdnInstanceSslBindRequest struct {
	CertId  int64    `json:"cert_id"`
	Domains []string `json:"domains"`
}

type RcdnInstanceSslBindResponse struct {
	sdkResponseBase
}

func (c *Client) RcdnInstanceSslBind(instanceId int64, req *RcdnInstanceSslBindRequest) (*RcdnInstanceSslBindResponse, error) {
	return c.RcdnInstanceSslBindWithContext(context.Background(), instanceId, req)
}

func (c *Client) RcdnInstanceSslBindWithContext(ctx context.Context, instanceId int64, req *RcdnInstanceSslBindRequest) (*RcdnInstanceSslBindResponse, error) {
	if instanceId == 0 {
		return nil, fmt.Errorf("sdkerr: bad request: unset instanceId")
	}

	httpreq, err := c.newRequest(http.MethodPost, fmt.Sprintf("/product/rcdn/instance/%d/ssl_bind", instanceId))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &RcdnInstanceSslBindResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
