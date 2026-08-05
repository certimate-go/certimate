package axisnow

import (
	"context"
	"net/http"
)

type AddCertificateRequest struct {
	Name        *string `json:"name,omitempty"`
	Type        *string `json:"type,omitempty"`
	Certificate *string `json:"certificate,omitempty"`
	PrivateKey  *string `json:"private_key,omitempty"`
}

type AddCertificateResponse struct {
	sdkResponseBase

	Result *Certificate `json:"result,omitempty"`
}

func (c *Client) AddCertificate(req *AddCertificateRequest) (*AddCertificateResponse, error) {
	return c.AddCertificateWithContext(context.Background(), req)
}

func (c *Client) AddCertificateWithContext(ctx context.Context, req *AddCertificateRequest) (*AddCertificateResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/certificates")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &AddCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
