package bunny

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type AddCustomCertificateRequest struct {
	Hostname       string `json:"Hostname"`
	Certificate    string `json:"Certificate"`
	CertificateKey string `json:"CertificateKey"`
}

func (c *Client) AddCustomCertificate(pullZoneId string, req *AddCustomCertificateRequest) error {
	return c.AddCustomCertificateWithContext(context.Background(), pullZoneId, req)
}

func (c *Client) AddCustomCertificateWithContext(ctx context.Context, pullZoneId string, req *AddCustomCertificateRequest) error {
	if pullZoneId == "" {
		return fmt.Errorf("sdkerr: bad request: unset pullZoneId")
	}

	httpreq, err := c.newRequest(http.MethodPost, fmt.Sprintf("/pullzone/%s/addCertificate", url.PathEscape(pullZoneId)))
	if err != nil {
		return err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	if _, err := c.doRequest(httpreq); err != nil {
		return err
	}

	return nil
}
