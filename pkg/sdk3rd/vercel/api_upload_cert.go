package vercel

import (
	"context"
	"net/http"
)

type UploadCertParams struct {
	CA             string `json:"ca"`
	Cert           string `json:"cert"`
	Key            string `json:"key"`
	SkipValidation bool   `json:"skipValidation,omitempty"`
}

type UploadCertResponse struct {
	sdkResponseBase
	ID        string `json:"id,omitempty"`
	ExpiresAt string `json:"expiresAt,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}

func (c *Client) UploadCert(req *UploadCertParams) (*UploadCertResponse, error) {
	return c.UploadCertWithContext(context.Background(), req)
}

func (c *Client) UploadCertWithContext(ctx context.Context, req *UploadCertParams) (*UploadCertResponse, error) {
	httpreq, err := c.newRequest(http.MethodPut, "/certs")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &UploadCertResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
