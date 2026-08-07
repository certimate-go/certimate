package samwaf

import (
	"context"
	"net/http"
)

type VipConfigRestartManagerResponse struct {
	sdkResponseBase
}

func (c *Client) VipConfigRestartManager() (*VipConfigRestartManagerResponse, error) {
	return c.VipConfigRestartManagerWithContext(context.Background())
}

func (c *Client) VipConfigRestartManagerWithContext(ctx context.Context) (*VipConfigRestartManagerResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/vipconfig/restartManager")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &VipConfigRestartManagerResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
