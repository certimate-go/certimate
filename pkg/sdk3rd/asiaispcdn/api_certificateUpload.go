package asiaispcdn

import (
	"context"
	"encoding/json"
	"net/http"
)

type CertificateUploadRequest struct {
	PublicKey  *string `json:"publicKey,omitempty"`
	PrivateKey *string `json:"privateKey,omitempty"`
	Name       *string `json:"name,omitempty"`
}

type CertificateUploadResponse struct {
	sdkResponseBase

	Data json.Number `json:"data,omitempty"`
}

func (c *Client) CertificateUpload(req *CertificateUploadRequest) (*CertificateUploadResponse, error) {
	return c.CertificateUploadWithContext(context.Background(), req)
}

func (c *Client) CertificateUploadWithContext(ctx context.Context, req *CertificateUploadRequest) (*CertificateUploadResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, req, "certificateUpload")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &CertificateUploadResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
