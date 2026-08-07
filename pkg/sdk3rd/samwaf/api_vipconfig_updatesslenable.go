package samwaf

import (
	"context"
	"net/http"
)

type VipConfigUpdateSslEnableRequest struct {
	SslEnable bool `json:"ssl_enable"`
}

type VipConfigUpdateSslEnableResponse struct {
	sdkResponseBase
}

func (c *Client) VipConfigUpdateSslEnable(req *VipConfigUpdateSslEnableRequest) (*VipConfigUpdateSslEnableResponse, error) {
	return c.VipConfigUpdateSslEnableWithContext(context.Background(), req)
}

func (c *Client) VipConfigUpdateSslEnableWithContext(ctx context.Context, req *VipConfigUpdateSslEnableRequest) (*VipConfigUpdateSslEnableResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/vipconfig/updateSslEnable")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &VipConfigUpdateSslEnableResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
