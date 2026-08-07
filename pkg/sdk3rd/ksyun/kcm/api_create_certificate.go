package kcm

import (
	"context"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type CreateCertificateRequest struct {
	Region           *string `json:"Region,omitempty"           url:"Region,omitempty"`
	CertificateName  *string `json:"CertificateName,omitempty"  url:"CertificateName,omitempty"`
	Description      *string `json:"Description,omitempty"      url:"Description,omitempty"`
	PublicKey        *string `json:"PublicKey,omitempty"        url:"-"`
	PrivateKey       *string `json:"PrivateKey,omitempty"       url:"-"`
	CertificateType  *string `json:"CertificateType,omitempty"  url:"CertificateType,omitempty"`
	Source           *string `json:"Source,omitempty"           url:"Source,omitempty"`
	SSLCertificateId *string `json:"SslCertificateId,omitempty" url:"SslCertificateId,omitempty"`
}

type CreateCertificateResponse struct {
	sdkResponseBase

	Certificate *LBCertificate `json:"Certificate,omitempty"`
}

func (c *Client) CreateCertificate(req *CreateCertificateRequest) (*CreateCertificateResponse, error) {
	return c.CreateCertificateWithContext(context.Background(), req)
}

func (c *Client) CreateCertificateWithContext(ctx context.Context, req *CreateCertificateRequest) (*CreateCertificateResponse, error) {
	params := &struct {
		CreateCertificateRequest `json:",inline"`
		Action                   string
		Version                  string
	}{
		CreateCertificateRequest: *req,
		Action:                   "CreateCertificate",
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

	result := &CreateCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
