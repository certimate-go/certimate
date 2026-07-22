package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type UploadServerCertificateRequest = iam.UploadServerCertificateInput

type UploadServerCertificateResponse = iam.UploadServerCertificateOutput

func (c *Client) UploadServerCertificate(req *UploadServerCertificateRequest) (*UploadServerCertificateResponse, error) {
	return c.UploadServerCertificateWithContext(context.Background(), req)
}

func (c *Client) UploadServerCertificateWithContext(ctx context.Context, req *UploadServerCertificateRequest) (*UploadServerCertificateResponse, error) {
	params := &struct {
		UploadServerCertificateRequest `json:",inline"`
		Action                         string
		Version                        string
	}{
		UploadServerCertificateRequest: *req,
		Action:                         "UploadServerCertificate",
		Version:                        "2010-05-08",
	}

	httpreq, err := c.newRequest(params)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &UploadServerCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
