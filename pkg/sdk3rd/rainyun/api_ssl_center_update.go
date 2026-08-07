package rainyun

import (
	"context"
	"fmt"
	"net/http"
)

type SslCenterUpdateRequest struct {
	Cert string `json:"cert"`
	Key  string `json:"key"`
}

type SslCenterUpdateResponse struct {
	sdkResponseBase
}

func (c *Client) SslCenterUpdate(certId int64, req *SslCenterUpdateRequest) (*SslCenterUpdateResponse, error) {
	return c.SslCenterUpdateWithContext(context.Background(), certId, req)
}

func (c *Client) SslCenterUpdateWithContext(ctx context.Context, certId int64, req *SslCenterUpdateRequest) (*SslCenterUpdateResponse, error) {
	if certId == 0 {
		return nil, fmt.Errorf("sdkerr: bad request: unset certId")
	}

	httpreq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/product/sslcenter/%d", certId))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &SslCenterUpdateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
