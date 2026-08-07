package kcm

import (
	"context"
	"net/http"
)

type UploadCertificateRequest struct {
	ProjectId *int64  `json:"ProjectId,omitempty"`
	CertName  *string `json:"CertName,omitempty"`
	CertFile  *string `json:"CertFile,omitempty"`
	CertKey   *string `json:"CertKey,omitempty"`
}

type UploadCertificateResponse struct {
	sdkResponseBase

	Success bool             `json:"Success"`
	Ret     *UserCertificate `json:"Ret,omitempty"`
}

func (c *Client) UploadCertificate(req *UploadCertificateRequest) (*UploadCertificateResponse, error) {
	return c.UploadCertificateWithContext(context.Background(), req)
}

func (c *Client) UploadCertificateWithContext(ctx context.Context, req *UploadCertificateRequest) (*UploadCertificateResponse, error) {
	params := &struct {
		UploadCertificateRequest `json:",inline"`
		Action                   string
		Version                  string
	}{
		UploadCertificateRequest: *req,
		Action:                   "UploadCertificate",
		Version:                  "2016-03-04",
	}

	httpreq, err := c.newRequest(http.MethodPost, "/", params)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &UploadCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
