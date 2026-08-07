package kcm

import (
	"context"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type ModifyCertificateRequest struct {
	Region           *string `json:"Region,omitempty"           url:"Region,omitempty"`
	CertificateId    *string `json:"CertificateId,omitempty"    url:"CertificateId,omitempty"`
	CertificateName  *string `json:"CertificateName,omitempty"  url:"CertificateName,omitempty"`
	Description      *string `json:"Description,omitempty"      url:"Description,omitempty"`
	PublicKey        *string `json:"PublicKey,omitempty"        url:"-"`
	PrivateKey       *string `json:"PrivateKey,omitempty"       url:"-"`
	SSLCertificateId *string `json:"SslCertificateId,omitempty" url:"SslCertificateId,omitempty"`
}

type ModifyCertificateResponse struct {
	sdkResponseBase

	Certificate *LBCertificate `json:"Certificate,omitempty"`
}

func (c *Client) ModifyCertificate(req *ModifyCertificateRequest) (*ModifyCertificateResponse, error) {
	return c.ModifyCertificateWithContext(context.Background(), req)
}

func (c *Client) ModifyCertificateWithContext(ctx context.Context, req *ModifyCertificateRequest) (*ModifyCertificateResponse, error) {
	params := &struct {
		ModifyCertificateRequest `json:",inline"`
		Action                   string
		Version                  string
	}{
		ModifyCertificateRequest: *req,
		Action:                   "ModifyCertificate",
		Version:                  "2016-03-04",
	}

	httpreq, err := c.newRequest(http.MethodPost, "/", params)
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

	result := &ModifyCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
