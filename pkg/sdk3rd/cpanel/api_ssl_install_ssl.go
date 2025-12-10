package baishan

import (
	"context"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type SSLInstallSSLRequest struct {
	Domain   *string `url:"domain,omitempty"`
	Cert     *string `url:"cert,omitempty"`
	Key      *string `url:"key,omitempty"`
	CABundle *string `url:"cabundle,omitempty"`
}

type SSLInstallSSLResponse struct {
	apiResponseBase

	Data *struct {
		User                    string   `json:"user"`
		Domain                  string   `json:"domain"`
		ExtraCertificateDomains []string `json:"extra_certificate_domains,omitempty"`
		WarningDomains          []string `json:"warning_domains,omitempty"`
		WorkingDomains          []string `json:"working_domains,omitempty"`
		CertId                  string   `json:"cert_id"`
		KeyId                   string   `json:"key_id"`
	} `json:"data,omitempty"`
}

func (c *Client) SSLInstallSSL(req *SSLInstallSSLRequest) (*SSLInstallSSLResponse, error) {
	return c.SSLInstallSSLWithContext(context.Background(), req)
}

func (c *Client) SSLInstallSSLWithContext(ctx context.Context, req *SSLInstallSSLRequest) (*SSLInstallSSLResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, "/SSL/install_ssl")
	if err != nil {
		return nil, err
	} else {
		values, err := qs.Values(req)
		if err != nil {
			return nil, err
		}

		httpreq.SetQueryParamsFromValues(values)
		httpreq.SetContext(ctx)
	}

	result := &SSLInstallSSLResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
