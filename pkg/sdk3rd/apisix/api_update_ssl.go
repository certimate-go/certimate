package apisix

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type UpdateSSLRequest Certificate

type UpdateSSLResponse Certificate

func (c *Client) UpdateSSL(sslId string, req *UpdateSSLRequest) (*UpdateSSLResponse, error) {
	return c.UpdateSSLWithContext(context.Background(), sslId, req)
}

func (c *Client) UpdateSSLWithContext(ctx context.Context, sslId string, req *UpdateSSLRequest) (*UpdateSSLResponse, error) {
	if sslId == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset sslId")
	}

	httpreq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/ssls/%s", url.PathEscape(sslId)))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &UpdateSSLResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
