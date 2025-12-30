package faas

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type UpdateCustomDomainRequest struct {
	RegionId    *string                  `json:"-"`
	DomainName  *string                  `json:"domainName,omitempty"`
	Protocol    *string                  `json:"protocol,omitempty"`
	AuthConfig  *CustomDomainAuthConfig  `json:"authConfig,omitempty"`
	CertConfig  *CustomDomainCertConfig  `json:"certConfig,omitempty"`
	RouteConfig *CustomDomainRouteConfig `json:"routeConfig,omitempty"`
}

type UpdateCustomDomainResponse struct {
	apiResponseBase

	ReturnObj *CustomDomainRecord `json:"returnObj,omitempty"`
}

func (c *Client) UpdateCustomDomain(req *UpdateCustomDomainRequest) (*UpdateCustomDomainResponse, error) {
	return c.UpdateCustomDomainWithContext(context.Background(), req)
}

func (c *Client) UpdateCustomDomainWithContext(ctx context.Context, req *UpdateCustomDomainRequest) (*UpdateCustomDomainResponse, error) {
	httpreq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/openapi/v1/domains/customdomains/%s", url.PathEscape(*req.DomainName)))
	if err != nil {
		return nil, err
	} else {
		if req.RegionId != nil {
			httpreq.SetHeader("regionId", *req.RegionId)
		}

		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &UpdateCustomDomainResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
