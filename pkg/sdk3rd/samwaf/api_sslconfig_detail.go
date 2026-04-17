package samwaf

import (
	"context"
	"fmt"
	"net/http"
)

type SslConfigDetailResponse struct {
	sdkResponseBase
	Data *SslConfig `json:"data,omitempty"`
}

func (c *Client) SslConfigDetail(sslId string) (*SslConfigDetailResponse, error) {
	return c.SslConfigDetailWithContext(context.Background(), sslId)
}

func (c *Client) SslConfigDetailWithContext(ctx context.Context, sslId string) (*SslConfigDetailResponse, error) {
	if sslId == "" {
		return nil, fmt.Errorf("sdkerr: unset sslId")
	}

	httpreq, err := c.newRequest(http.MethodGet, "/sslconfig/detail")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetQueryParam("id", sslId)
		httpreq.SetContext(ctx)
	}

	result := &SslConfigDetailResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
