package cdn

import (
	"context"
	"net/http"
)

type ConfigCertificateRequest struct {
	Enable            *string `json:"Enable,omitempty"`
	DomainIds         *string `json:"DomainIds,omitempty"`
	CertificateId     *string `json:"CertificateId,omitempty"`
	CertificateName   *string `json:"CertificateName,omitempty"`
	ServerCertificate *string `json:"ServerCertificate,omitempty"`
	PrivateKey        *string `json:"PrivateKey,omitempty"`
}

type ConfigCertificateResponse struct {
	sdkResponseBase

	CertificateId string `json:"CertificateId,omitempty"`
}

func (c *Client) ConfigCertificate(req *ConfigCertificateRequest) (*ConfigCertificateResponse, error) {
	return c.ConfigCertificateWithContext(context.Background(), req)
}

func (c *Client) ConfigCertificateWithContext(ctx context.Context, req *ConfigCertificateRequest) (*ConfigCertificateResponse, error) {
	params := &struct {
		ConfigCertificateRequest `json:",inline"`
		Action                   string
		Version                  string
	}{
		ConfigCertificateRequest: *req,
		Action:                   "ConfigCertificate",
		Version:                  "2016-09-01",
	}

	httpreq, err := c.newRequest(http.MethodPost, "/2016-09-01/cert/ConfigCertificate", params)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &ConfigCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
