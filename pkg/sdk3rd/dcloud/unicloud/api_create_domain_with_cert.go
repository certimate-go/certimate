package unicloud

import (
	"context"
	"net/http"
)

type CreateDomainWithCertRequest struct {
	Provider string `json:"provider"`
	SpaceId  string `json:"spaceId"`
	Domain   string `json:"domain"`
	Cert     string `json:"cert"`
	Key      string `json:"key"`
}

type CreateDomainWithCertResponse struct {
	sdkResponseBase
}

func (c *Client) CreateDomainWithCert(req *CreateDomainWithCertRequest) (*CreateDomainWithCertResponse, error) {
	return c.CreateDomainWithCertWithContext(context.Background(), req)
}

func (c *Client) CreateDomainWithCertWithContext(ctx context.Context, req *CreateDomainWithCertRequest) (*CreateDomainWithCertResponse, error) {
	if err := c.ensureApiUserToken(ctx); err != nil {
		return nil, err
	}

	resp := &CreateDomainWithCertResponse{}
	err := c.sendRequestWithResult(ctx, http.MethodPost, "/host/create-domain-with-cert", req, resp)
	return resp, err
}
