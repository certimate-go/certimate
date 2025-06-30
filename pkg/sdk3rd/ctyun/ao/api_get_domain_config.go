package ao

import (
	"context"
	"net/http"
)

type GetDomainConfigRequest struct {
	Domain      *string `json:"domain,omitempty"`
	ProductCode *string `json:"product_code,omitempty"`
}

type GetDomainConfigResponse struct {
	apiResponseBase

	ReturnObj *struct {
		Domain      string                          `json:"domain"`
		ProductCode string                          `json:"product_code"`
		Status      int32                           `json:"status"`
		AreaScope   int32                           `json:"area_scope"`
		Cname       string                          `json:"cname"`
		Origin      []*DomainOriginConfigWithWeight `json:"origin,omitempty"`
		HttpsStatus string                          `json:"https_status"`
		HttpsBasic  *DomainHttpsBasicConfig         `json:"https_basic,omitempty"`
		CertName    string                          `json:"cert_name"`
	} `json:"returnObj,omitempty"`
}

func (c *Client) GetDomainConfig(req *GetDomainConfigRequest) (*GetDomainConfigResponse, error) {
	return c.GetDomainConfigWithContext(context.Background(), req)
}

func (c *Client) GetDomainConfigWithContext(ctx context.Context, req *GetDomainConfigRequest) (*GetDomainConfigResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/ctapi/v1/accessone/domain/config")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &GetDomainConfigResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
