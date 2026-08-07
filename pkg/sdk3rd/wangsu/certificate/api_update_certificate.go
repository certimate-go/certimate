package certificate

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type UpdateCertificateRequest struct {
	Name        *string `json:"name,omitempty"`
	Certificate *string `json:"certificate,omitempty"`
	PrivateKey  *string `json:"privateKey,omitempty"`
	Comment     *string `json:"comment,omitempty" `
}

type UpdateCertificateResponse struct {
	sdkResponseBase
}

func (c *Client) UpdateCertificate(certificateId string, req *UpdateCertificateRequest) (*UpdateCertificateResponse, error) {
	return c.UpdateCertificateWithContext(context.Background(), certificateId, req)
}

func (c *Client) UpdateCertificateWithContext(ctx context.Context, certificateId string, req *UpdateCertificateRequest) (*UpdateCertificateResponse, error) {
	if certificateId == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset certificateId")
	}

	httpreq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/certificate/%s", url.PathEscape(certificateId)))
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
