package acm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/acm"
)

type ImportCertificateRequest = acm.ImportCertificateInput

type ImportCertificateResponse = acm.ImportCertificateOutput

func (c *Client) ImportCertificate(req *ImportCertificateRequest) (*ImportCertificateResponse, error) {
	return c.ImportCertificateWithContext(context.Background(), req)
}

func (c *Client) ImportCertificateWithContext(ctx context.Context, req *ImportCertificateRequest) (*ImportCertificateResponse, error) {
	httpreq, err := c.newRequest(buildAmzTarget("ImportCertificate"))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &ImportCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
