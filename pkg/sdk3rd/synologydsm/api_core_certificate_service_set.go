package synologydsm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type ServiceCertificateSetting struct {
	Service   *CertificateService `json:"service"`
	OldCertID string              `json:"old_id"`
	CertID    string              `json:"id"`
}

type SetServiceCertificateRequest struct {
	Settings []*ServiceCertificateSetting `json:"settings"`
}

type SetServiceCertificateResponse struct {
	sdkResponseBase
}

func (c *Client) SetServiceCertificate(req *SetServiceCertificateRequest) (*SetServiceCertificateResponse, error) {
	bsettings, _ := json.Marshal(req.Settings)
	params := url.Values{
		"api":      {"SYNO.Core.Certificate.Service"},
		"method":   {"set"},
		"version":  {"1"},
		"settings": {string(bsettings)},
	}

	httpreq, err := c.newRequest(http.MethodPost, fmt.Sprintf("/webapi/entry.cgi?_sid=%s", c.sid))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetHeader("Content-Type", "application/x-www-form-urlencoded")
		httpreq.SetFormDataFromValues(params)
	}

	result := &SetServiceCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
