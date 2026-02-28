package flyio

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type ImportCustomCertificateRequest struct {
	AppName    string `json:"-"`
	Hostname   string `json:"hostname"`
	Fullchain  string `json:"fullchain"`
	PrivateKey string `json:"private_key"`
}

type ImportCustomCertificateResponse struct {
	sdkResponseBase

	Hostname     string `json:"hostname"`
	Configured   bool   `json:"configured"`
	Status       string `json:"status"`
	Certificates []*struct {
		Source    string `json:"source"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
		ExpiresAt string `json:"expires_at"`
		Issuer    string `json:"issuer"`
	} `json:"certificates"`
}

func (c *Client) ImportCustomCertificate(req *ImportCustomCertificateRequest) (*ImportCustomCertificateResponse, error) {
	return c.ImportCustomCertificateWithContext(context.Background(), req)
}

func (c *Client) ImportCustomCertificateWithContext(ctx context.Context, req *ImportCustomCertificateRequest) (*ImportCustomCertificateResponse, error) {
	if req.AppName == "" {
		return nil, fmt.Errorf("sdkerr: unset appName")
	}

	path := fmt.Sprintf("/apps/%s/certificates/custom", url.PathEscape(req.AppName))
	httpreq, err := c.newRequest(http.MethodPost, path)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &ImportCustomCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
