package cdn

import (
	"context"
	"net/http"
)

type SetCertificateRequest struct {
	CertificateId     *string `json:"CertificateId,omitempty"`
	CertificateName   *string `json:"CertificateName,omitempty"`
	ServerCertificate *string `json:"ServerCertificate,omitempty"`
	PrivateKey        *string `json:"PrivateKey,omitempty"`
}

type SetCertificateResponse struct {
	sdkResponseBase

	CertificateId string `json:"CertificateId,omitempty"`
}

func (c *Client) SetCertificate(req *SetCertificateRequest) (*SetCertificateResponse, error) {
	return c.SetCertificateWithContext(context.Background(), req)
}

func (c *Client) SetCertificateWithContext(ctx context.Context, req *SetCertificateRequest) (*SetCertificateResponse, error) {
	params := &struct {
		SetCertificateRequest `json:",inline"`
		Action                string
		Version               string
	}{
		SetCertificateRequest: *req,
		Action:                "SetCertificate",
		Version:               "2016-09-01",
	}

	httpreq, err := c.newRequest(http.MethodPost, "/2016-09-01/cert/SetCertificate", params)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &SetCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
