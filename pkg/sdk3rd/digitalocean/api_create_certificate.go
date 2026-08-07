package digitalocean

import (
	"context"
	"net/http"
)

type CreateCertificateRequest struct {
	Name             *string   `json:"name,omitempty"`
	DNSNames         []*string `json:"dns_names,omitempty"`
	Type             *string   `json:"type,omitempty"`
	CertificateChain *string   `json:"certificate_chain,omitempty"`
	LeafCertificate  *string   `json:"leaf_certificate,omitempty"`
	PrivateKey       *string   `json:"private_key,omitempty"`
}

type CreateCertificateResponse struct {
	sdkResponseBase

	Certificate *Certificate `json:"certificate,omitempty"`
}

func (c *Client) CreateCertificate(req *CreateCertificateRequest) (*CreateCertificateResponse, error) {
	return c.CreateCertificateWithContext(context.Background(), req)
}

func (c *Client) CreateCertificateWithContext(ctx context.Context, req *CreateCertificateRequest) (*CreateCertificateResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/certificates")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &CreateCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
