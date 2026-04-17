package samwaf

import (
	"context"
	"net/http"
)

type SslConfigEditRequest struct {
	Id          string `json:"id"`
	CertContent string `json:"cert_content"`
	CertPath    string `json:"cert_path"`
	KeyContent  string `json:"key_content"`
	KeyPath     string `json:"key_path"`
}

type SslConfigEditResponse struct {
	sdkResponseBase
}

func (c *Client) SslConfigEdit(req *SslConfigEditRequest) (*SslConfigEditResponse, error) {
	return c.SslConfigEditWithContext(context.Background(), req)
}

func (c *Client) SslConfigEditWithContext(ctx context.Context, req *SslConfigEditRequest) (*SslConfigEditResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/sslconfig/edit")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &SslConfigEditResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
