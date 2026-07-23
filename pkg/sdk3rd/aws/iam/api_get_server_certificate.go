package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type GetServerCertificateRequest = iam.GetServerCertificateInput

type GetServerCertificateResponse = iam.GetServerCertificateOutput

func (c *Client) GetServerCertificate(req *GetServerCertificateRequest) (*GetServerCertificateResponse, error) {
	return c.GetServerCertificateWithContext(context.Background(), req)
}

func (c *Client) GetServerCertificateWithContext(ctx context.Context, req *GetServerCertificateRequest) (*GetServerCertificateResponse, error) {
	params := &struct {
		GetServerCertificateRequest `json:",inline"`
		Action                      string
		Version                     string
	}{
		GetServerCertificateRequest: *req,
		Action:                      "GetServerCertificate",
		Version:                     "2010-05-08",
	}

	httpreq, err := c.newRequest(params)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &GetServerCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
