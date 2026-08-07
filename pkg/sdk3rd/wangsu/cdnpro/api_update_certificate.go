package cdnpro

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type UpdateCertificateRequest struct {
	Timestamp   int64               `json:"-"`
	Name        *string             `json:"name,omitempty"`
	Description *string             `json:"description,omitempty"`
	AutoRenew   *string             `json:"autoRenew,omitempty"`
	ForceRenew  *bool               `json:"forceRenew,omitempty"`
	NewVersion  *CertificateVersion `json:"newVersion,omitempty"`
}

type UpdateCertificateResponse struct {
	sdkResponseBase

	CertificateLocation string `json:"location,omitempty"`
}

func (c *Client) UpdateCertificate(certificateId string, req *UpdateCertificateRequest) (*UpdateCertificateResponse, error) {
	return c.UpdateCertificateWithContext(context.Background(), certificateId, req)
}

func (c *Client) UpdateCertificateWithContext(ctx context.Context, certificateId string, req *UpdateCertificateRequest) (*UpdateCertificateResponse, error) {
	if certificateId == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset certificateId")
	}

	httpreq, err := c.newRequest(http.MethodPatch, fmt.Sprintf("/cdn/certificates/%s", url.PathEscape(certificateId)))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetHeader("X-CNC-Timestamp", fmt.Sprintf("%d", req.Timestamp))
		httpreq.SetContext(ctx)
	}

	result := &UpdateCertificateResponse{}
	if httpresp, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	} else {
		result.CertificateLocation = httpresp.Header().Get("Location")
	}

	return result, nil
}
