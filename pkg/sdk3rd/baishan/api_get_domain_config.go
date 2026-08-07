package baishan

import (
	"context"
	"net/http"
)

type GetDomainConfigRequest struct {
	Domains *string   `json:"domains,omitempty" url:"domains,omitempty"`
	Config  []*string `json:"config,omitempty"  url:"config,omitempty"`
}

type GetDomainConfigResponse struct {
	sdkResponseBase

	Data []*struct {
		Domain string        `json:"domain"`
		Config *DomainConfig `json:"config"`
	} `json:"data,omitempty"`
}

func (c *Client) GetDomainConfig(req *GetDomainConfigRequest) (*GetDomainConfigResponse, error) {
	return c.GetDomainConfigWithContext(context.Background(), req)
}

func (c *Client) GetDomainConfigWithContext(ctx context.Context, req *GetDomainConfigRequest) (*GetDomainConfigResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, "/v2/domain/config")
	if err != nil {
		return nil, err
	} else {
		if req.Domains != nil {
			httpreq.SetQueryParam("domains", *req.Domains)
		}
		if req.Config != nil {
			for _, config := range req.Config {
				if config != nil {
					httpreq.QueryParam.Add("config[]", *config)
				}
			}
		}

		httpreq.SetContext(ctx)
	}

	result := &GetDomainConfigResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
