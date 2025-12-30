package nginxproxymanager

import (
	"context"
	"net/http"
)

type NginxCreateCertificateRequest struct {
	Provider string `json:"provider"`
	NiceName string `json:"nice_name"`
}

type NginxCreateCertificateResponse struct {
	CertificateRecord
}

func (c *Client) NginxCreateCertificate(req *NginxCreateCertificateRequest) (*NginxCreateCertificateResponse, error) {
	return c.NginxCreateCertificateWithContext(context.Background(), req)
}

func (c *Client) NginxCreateCertificateWithContext(ctx context.Context, req *NginxCreateCertificateRequest) (*NginxCreateCertificateResponse, error) {
	if err := c.ensureJwtTokenExists(); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodPost, "/nginx/certificates")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &NginxCreateCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
