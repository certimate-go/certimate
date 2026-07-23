package amplify

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/service/amplify"
)

type UpdateDomainAssociationRequest = amplify.UpdateDomainAssociationInput

type UpdateDomainAssociationResponse = amplify.UpdateDomainAssociationOutput

func (c *Client) UpdateDomainAssociation(req *UpdateDomainAssociationRequest) (*UpdateDomainAssociationResponse, error) {
	return c.UpdateDomainAssociationWithContext(context.Background(), req)
}

func (c *Client) UpdateDomainAssociationWithContext(ctx context.Context, req *UpdateDomainAssociationRequest) (*UpdateDomainAssociationResponse, error) {
	if req.AppId == nil {
		return nil, fmt.Errorf("sdkerr: bad request: unset appId")
	}
	if req.DomainName == nil {
		return nil, fmt.Errorf("sdkerr: bad request: unset domainName")
	}

	path := fmt.Sprintf("/apps/%s/domains/%s", url.PathEscape(*req.AppId), url.PathEscape(*req.DomainName))
	httpreq, err := c.newRequest(http.MethodPost, path, req)
	if err != nil {
		return nil, err
	} else {
		if m, ok := httpreq.Body.(map[string]any); ok {
			delete(m, "appId")
			delete(m, "domainName")
		}

		httpreq.SetContext(ctx)
	}

	result := &UpdateDomainAssociationResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
