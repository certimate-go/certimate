package cdnpro

import (
	"context"
	"fmt"
	"net/http"
)

type CreateCertificateRequest struct {
	Timestamp   int64               `json:"-"`
	Name        *string             `json:"name,omitempty"`
	Description *string             `json:"description,omitempty"`
	AutoRenew   *string             `json:"autoRenew,omitempty"`
	ForceRenew  *bool               `json:"forceRenew,omitempty"`
	NewVersion  *CertificateVersion `json:"newVersion,omitempty"`
}

type CreateCertificateResponse struct {
	sdkResponseBase

	CertificateLocation string `json:"location,omitempty"`
}

func (c *Client) CreateCertificate(req *CreateCertificateRequest) (*CreateCertificateResponse, error) {
	return c.CreateCertificateWithContext(context.Background(), req)
}

func (c *Client) CreateCertificateWithContext(ctx context.Context, req *CreateCertificateRequest) (*CreateCertificateResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/cdn/certificates")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetHeader("X-CNC-Timestamp", fmt.Sprintf("%d", req.Timestamp))
		httpreq.SetContext(ctx)
	}

	result := &CreateCertificateResponse{}
	if httpresp, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	} else {
		result.CertificateLocation = httpresp.Header().Get("Location")
	}

	return result, nil
}
