package netlify

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type ProvisionSiteTLSCertificateRequest struct {
	Certificate    string `json:"certificate"`
	CACertificates string `json:"ca_certificates"`
	Key            string `json:"key"`
}

type ProvisionSiteTLSCertificateResponse struct {
	sdkResponseBase
	Domains   []string `json:"domains,omitempty"`
	State     string   `json:"state,omitempty"`
	ExpiresAt string   `json:"expires_at,omitempty"`
	CreatedAt string   `json:"created_at,omitempty"`
	UpdatedAt string   `json:"updated_at,omitempty"`
}

func (c *Client) ProvisionSiteTLSCertificate(siteId string, req *ProvisionSiteTLSCertificateRequest) (*ProvisionSiteTLSCertificateResponse, error) {
	return c.ProvisionSiteTLSCertificateWithContext(context.Background(), siteId, req)
}

func (c *Client) ProvisionSiteTLSCertificateWithContext(ctx context.Context, siteId string, req *ProvisionSiteTLSCertificateRequest) (*ProvisionSiteTLSCertificateResponse, error) {
	if siteId == "" {
		return nil, fmt.Errorf("sdkerr: unset siteId")
	}

	httpreq, err := c.newRequest(http.MethodPost, fmt.Sprintf("/sites/%s/ssl", url.PathEscape(siteId)))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetQueryParams(map[string]string{
			"certificate":     req.Certificate,
			"ca_certificates": req.CACertificates,
			"key":             req.Key,
		})
		httpreq.SetContext(ctx)
	}

	result := &ProvisionSiteTLSCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
