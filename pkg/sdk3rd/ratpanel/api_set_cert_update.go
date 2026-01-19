package ratpanel

import (
	"context"
	"fmt"
	"net/http"
)

type CertUpdateRequest struct {
	CertId      int64    `json:"id"`
	Type        string   `json:"type"`
	Domains     []string `json:"domains"`
	Certificate string   `json:"cert"`
	PrivateKey  string   `json:"key"`
}

type CertUpdateResponse struct {
	sdkResponseBase
}

func (c *Client) CertUpdate(req *CertUpdateRequest) (*CertUpdateResponse, error) {
	return c.CertUpdateWithContext(context.Background(), req)
}

func (c *Client) CertUpdateWithContext(ctx context.Context, req *CertUpdateRequest) (*CertUpdateResponse, error) {
	httpreq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/cert/cert/%d", req.CertId))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &CertUpdateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
