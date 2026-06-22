package console

import (
	"context"
	"net/http"
)

type MigrateHttpsDomainRequest struct {
	CertificateId string `json:"crt_id"`
	Domain        string `json:"domain_name"`
}

type MigrateHttpsDomainResponse struct {
	sdkResponseBase

	Data *struct {
		sdkResponseBaseData

		Status bool `json:"status"`
	} `json:"data,omitempty"`
}

func (c *Client) MigrateHttpsDomain(req *MigrateHttpsDomainRequest) (*MigrateHttpsDomainResponse, error) {
	return c.MigrateHttpsDomainWithContext(context.Background(), req)
}

func (c *Client) MigrateHttpsDomainWithContext(ctx context.Context, req *MigrateHttpsDomainRequest) (*MigrateHttpsDomainResponse, error) {
	if err := c.ensureCookies(ctx); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodPost, "/api/https/migrate/domain")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &MigrateHttpsDomainResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
