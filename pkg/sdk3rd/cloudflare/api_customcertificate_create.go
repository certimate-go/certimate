package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type CustomCertificateCreateRequest struct {
	ZoneId          string            `json:"-"`
	CustomCsrId     *string           `json:"custom_csr_id,omitempty"`
	Certificate     *string           `json:"certificate,omitempty"`
	PrivateKey      *string           `json:"private_key,omitempty"`
	BundleMethod    *string           `json:"bundle_method,omitempty"`
	Type            *string           `json:"type,omitempty"`
	Deploy          *string           `json:"deploy,omitempty"`
	Policy          *string           `json:"policy,omitempty"`
	GeoRestrictions []*GeoRestriction `json:"geo_restrictions,omitempty"`
}

type CustomCertificateCreateResponse struct {
	sdkResponseBase

	Result *CustomCertificate `json:"result,omitempty"`
}

func (c *Client) CustomCertificateCreate(req *CustomCertificateCreateRequest) (*CustomCertificateCreateResponse, error) {
	return c.CustomCertificateCreateWithContext(context.Background(), req)
}

func (c *Client) CustomCertificateCreateWithContext(ctx context.Context, req *CustomCertificateCreateRequest) (*CustomCertificateCreateResponse, error) {
	path := fmt.Sprintf("/zones/%s/custom_certificates", url.PathEscape(req.ZoneId))
	httpreq, err := c.newRequest(http.MethodPost, path)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &CustomCertificateCreateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
