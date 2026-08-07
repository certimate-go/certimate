package console

import (
	"context"
	"net/http"
)

type UpdateHttpsCertificateManagerRequest struct {
	CertificateId string `json:"certificate_id"`
	Domain        string `json:"domain"`
	Https         bool   `json:"https"`
	ForceHttps    bool   `json:"force_https"`
}

type UpdateHttpsCertificateManagerResponse struct {
	sdkResponseBase

	Data *struct {
		sdkResponseBaseData

		Status bool `json:"status"`
	} `json:"data,omitempty"`
}

func (c *Client) UpdateHttpsCertificateManager(req *UpdateHttpsCertificateManagerRequest) (*UpdateHttpsCertificateManagerResponse, error) {
	return c.UpdateHttpsCertificateManagerWithContext(context.Background(), req)
}

func (c *Client) UpdateHttpsCertificateManagerWithContext(ctx context.Context, req *UpdateHttpsCertificateManagerRequest) (*UpdateHttpsCertificateManagerResponse, error) {
	if err := c.ensureCookies(ctx); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodPost, "/api/https/certificate/manager")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &UpdateHttpsCertificateManagerResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
