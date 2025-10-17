package ao

import (
	"context"
	"net/http"
)

type ModifyDomainConfigRequest struct {
	Domain      *string                 `json:"domain,omitempty"`
	ProductCode *string                 `json:"product_code,omitempty"`
	Origin      []*DomainOriginConfig   `json:"origin,omitempty"`
	HttpsStatus *string                 `json:"https_status,omitempty"`
	HttpsBasic  *DomainHttpsBasicConfig `json:"https_basic,omitempty"`
	CertName    *string                 `json:"cert_name,omitempty"`
}

type ModifyDomainConfigResponse struct {
	apiResponseBase
}

func (c *Client) ModifyDomainConfig(req *ModifyDomainConfigRequest) (*ModifyDomainConfigResponse, error) {
	return c.ModifyDomainConfigWithContext(context.Background(), req)
}

func (c *Client) ModifyDomainConfigWithContext(ctx context.Context, req *ModifyDomainConfigRequest) (*ModifyDomainConfigResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/ctapi/v1/scdn/domain/modify_config")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &ModifyDomainConfigResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
