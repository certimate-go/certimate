package apigatewayv2

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
)

type UpdateDomainNameRequest = apigatewayv2.UpdateDomainNameInput

type UpdateDomainNameResponse = apigatewayv2.UpdateDomainNameOutput

func (c *Client) UpdateDomainName(req *UpdateDomainNameRequest) (*UpdateDomainNameResponse, error) {
	return c.UpdateDomainNameWithContext(context.Background(), req)
}

func (c *Client) UpdateDomainNameWithContext(ctx context.Context, req *UpdateDomainNameRequest) (*UpdateDomainNameResponse, error) {
	if req.DomainName == nil {
		return nil, fmt.Errorf("sdkerr: bad request: unset domainName")
	}

	path := fmt.Sprintf("/domainnames/%s", url.PathEscape(*req.DomainName))
	httpreq, err := c.newRequest(http.MethodPatch, path, req)
	if err != nil {
		return nil, err
	} else {
		if m, ok := httpreq.Body.(map[string]any); ok {
			delete(m, "domainName")
		}

		httpreq.SetContext(ctx)
	}

	result := &UpdateDomainNameResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
