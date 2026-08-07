package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type CustomCertificateEditRequest struct {
	ZoneId          string            `json:"-"`
	CertificateId   string            `json:"-"`
	CustomCsrId     *string           `json:"custom_csr_id,omitempty"`
	Certificate     *string           `json:"certificate,omitempty"`
	PrivateKey      *string           `json:"private_key,omitempty"`
	BundleMethod    *string           `json:"bundle_method,omitempty"`
	Deploy          *string           `json:"deploy,omitempty"`
	Policy          *string           `json:"policy,omitempty"`
	GeoRestrictions []*GeoRestriction `json:"geo_restrictions,omitempty"`
}

type CustomCertificateEditResponse struct {
	sdkResponseBase

	Result *CustomCertificate `json:"result,omitempty"`
}

func (c *Client) CustomCertificateEdit(req *CustomCertificateEditRequest) (*CustomCertificateEditResponse, error) {
	return c.CustomCertificateEditWithContext(context.Background(), req)
}

func (c *Client) CustomCertificateEditWithContext(ctx context.Context, req *CustomCertificateEditRequest) (*CustomCertificateEditResponse, error) {
	path := fmt.Sprintf("/zones/%s/custom_certificates/%s", url.PathEscape(req.ZoneId), url.PathEscape(req.CertificateId))
	httpreq, err := c.newRequest(http.MethodPatch, path)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &CustomCertificateEditResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
