package master

import (
	"context"
	"fmt"
	"net/http"
)

type UpdateCertificateRequest struct {
	ClientId    int64  `json:"client_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	SSLPEM      string `json:"ssl_pem"`
	SSLKey      string `json:"ssl_key"`
	AutoRenewal bool   `json:"auto_renewal"`
}

type UpdateCertificateResponse struct {
	sdkResponseBase
}

func (c *Client) UpdateCertificate(certId int64, req *UpdateCertificateRequest) (*UpdateCertificateResponse, error) {
	return c.UpdateCertificateWithContext(context.Background(), certId, req)
}

func (c *Client) UpdateCertificateWithContext(ctx context.Context, certId int64, req *UpdateCertificateRequest) (*UpdateCertificateResponse, error) {
	if certId == 0 {
		return nil, fmt.Errorf("sdkerr: bad request: unset certId")
	}

	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/certificate/%d", certId))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &UpdateCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
