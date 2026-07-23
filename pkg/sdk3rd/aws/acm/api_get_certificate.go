package acm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/acm"
)

type GetCertificateRequest = acm.GetCertificateInput

type GetCertificateResponse = acm.GetCertificateOutput

func (c *Client) GetCertificate(req *GetCertificateRequest) (*GetCertificateResponse, error) {
	return c.GetCertificateWithContext(context.Background(), req)
}

func (c *Client) GetCertificateWithContext(ctx context.Context, req *GetCertificateRequest) (*GetCertificateResponse, error) {
	httpreq, err := c.newRequest(buildAmzTarget("GetCertificate"))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &GetCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
