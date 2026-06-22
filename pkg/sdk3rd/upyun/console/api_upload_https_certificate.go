package console

import (
	"context"
	"net/http"
)

type UploadHttpsCertificateRequest struct {
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"private_key"`
}

type UploadHttpsCertificateResponse struct {
	sdkResponseBase

	Data *struct {
		sdkResponseBaseData

		Status int32 `json:"status"`
		Result struct {
			CertificateId string `json:"certificate_id"`
			CommonName    string `json:"commonName"`
			Serial        string `json:"serial"`
		} `json:"result"`
	} `json:"data,omitempty"`
}

func (c *Client) UploadHttpsCertificate(req *UploadHttpsCertificateRequest) (*UploadHttpsCertificateResponse, error) {
	return c.UploadHttpsCertificateWithContext(context.Background(), req)
}

func (c *Client) UploadHttpsCertificateWithContext(ctx context.Context, req *UploadHttpsCertificateRequest) (*UploadHttpsCertificateResponse, error) {
	if err := c.ensureCookies(ctx); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodPost, "/api/https/certificate/")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &UploadHttpsCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
