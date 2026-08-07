package cdnfly

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type UpdateCertRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"des,omitempty"`
	Type        *string `json:"type,omitempty"`
	Cert        *string `json:"cert,omitempty"`
	Key         *string `json:"key,omitempty"`
	Enable      *bool   `json:"enable,omitempty"`
}

type UpdateCertResponse struct {
	sdkResponseBase
}

func (c *Client) UpdateCert(certId string, req *UpdateCertRequest) (*UpdateCertResponse, error) {
	return c.UpdateCertWithContext(context.Background(), certId, req)
}

func (c *Client) UpdateCertWithContext(ctx context.Context, certId string, req *UpdateCertRequest) (*UpdateCertResponse, error) {
	if certId == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset certId")
	}

	httpreq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/certs/%s", url.PathEscape(certId)))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &UpdateCertResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
