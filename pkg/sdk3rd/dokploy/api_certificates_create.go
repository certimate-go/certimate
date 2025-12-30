package dokploy

import (
	"context"
	"net/http"
)

type CertificatesCreateRequest struct {
	CertificateId   *string `json:"certificateId,omitempty"`
	Name            *string `json:"name,omitempty"`
	CertificateData *string `json:"certificateData,omitempty"`
	PrivateKey      *string `json:"privateKey,omitempty"`
	OrganizationId  *string `json:"organizationId,omitempty"`
	ServerId        *string `json:"serverId,omitempty"`
}

type CertificatesCreateResponse = Certificate

func (c *Client) CertificatesCreate(req *CertificatesCreateRequest) (*CertificatesCreateResponse, error) {
	return c.CertificatesCreateWithContext(context.Background(), req)
}

func (c *Client) CertificatesCreateWithContext(ctx context.Context, req *CertificatesCreateRequest) (*CertificatesCreateResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/certificates.create")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &CertificatesCreateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
