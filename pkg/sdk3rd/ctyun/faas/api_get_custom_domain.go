package faas

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	qs "github.com/google/go-querystring/query"
)

type GetCustomDomainRequest struct {
	RegionId   *string `json:"-" url:"-"`
	DomainName *string `json:"domainName,omitempty" url:"-"`
	CnameCheck *bool   `json:"cnameCheck,omitempty" url:"cnameCheck,omitempty"`
}

type GetCustomDomainResponse struct {
	apiResponseBase

	ReturnObj *CustomDomainRecord `json:"returnObj,omitempty"`
}

func (c *Client) GetCustomDomain(req *GetCustomDomainRequest) (*GetCustomDomainResponse, error) {
	return c.GetCustomDomainWithContext(context.Background(), req)
}

func (c *Client) GetCustomDomainWithContext(ctx context.Context, req *GetCustomDomainRequest) (*GetCustomDomainResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/openapi/v1/domains/customdomains/%s", url.PathEscape(*req.DomainName)))
	if err != nil {
		return nil, err
	} else {
		if req.RegionId != nil {
			httpreq.SetHeader("regionId", *req.RegionId)
		}

		values, err := qs.Values(req)
		if err != nil {
			return nil, err
		}

		httpreq.SetQueryParamsFromValues(values)
		httpreq.SetContext(ctx)
	}

	result := &GetCustomDomainResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
