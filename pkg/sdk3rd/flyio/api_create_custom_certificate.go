package flyio

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type CreateCustomCertificateRequest struct {
	AppName    string `json:"-"`
	Hostname   string `json:"hostname"`
	Fullchain  string `json:"fullchain"`
	PrivateKey string `json:"private_key"`
}

type CreateCustomCertificateResponse struct {
	sdkResponseBase

	Hostname     string              `json:"hostname"`
	Configured   bool                `json:"configured"`
	Status       string              `json:"status"`
	Certificates []CertificateDetail `json:"certificates"`
}

type CertificateDetail struct {
	Source string `json:"source"`
	Status string `json:"status"`
}

func (c *Client) CreateCustomCertificate(req *CreateCustomCertificateRequest) (*CreateCustomCertificateResponse, error) {
	return c.CreateCustomCertificateWithContext(context.Background(), req)
}

func (c *Client) CreateCustomCertificateWithContext(ctx context.Context, req *CreateCustomCertificateRequest) (*CreateCustomCertificateResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("sdkerr: nil request")
	}
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

	result := &CreateCustomCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
