package console

import (
	"context"
	"fmt"
	"net/http"
)

type HttpsCertificateManagerDomain struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	BucketId   int64  `json:"bucket_id"`
	BucketName string `json:"bucket_name"`
}

type GetHttpsCertificateManagerResponse struct {
	sdkResponseBase

	Data *struct {
		sdkResponseBaseData

		AuthenticateNum     int32                           `json:"authenticate_num"`
		AuthenticateDomains []string                        `json:"authenticate_domain"`
		Domains             []HttpsCertificateManagerDomain `json:"domains"`
	} `json:"data,omitempty"`
}

func (c *Client) GetHttpsCertificateManager(certificateId string) (*GetHttpsCertificateManagerResponse, error) {
	return c.GetHttpsCertificateManagerWithContext(context.Background(), certificateId)
}

func (c *Client) GetHttpsCertificateManagerWithContext(ctx context.Context, certificateId string) (*GetHttpsCertificateManagerResponse, error) {
	if certificateId == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset certificateId")
	}

	if err := c.ensureCookies(ctx); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodGet, "/api/https/certificate/manager/")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetQueryParam("certificate_id", certificateId)
		httpreq.SetContext(ctx)
	}

	result := &GetHttpsCertificateManagerResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
