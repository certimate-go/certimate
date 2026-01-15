package synologydsm

import (
	"net/http"
	"net/url"
)

type ListCertificatesResponse struct {
	sdkResponseBase
	Data *struct {
		Certificates []*CertificateInfo `json:"certificates"`
	} `json:"data,omitempty"`
}

func (c *Client) ListCertificates() (*ListCertificatesResponse, error) {
	params := url.Values{
		"api":     {"SYNO.Core.Certificate.CRT"},
		"method":  {"list"},
		"version": {"1"},
		"_sid":    {c.sid},
	}

	httpreq, err := c.newRequest(http.MethodPost, "/webapi/entry.cgi")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetHeader("Content-Type", "application/x-www-form-urlencoded")
		httpreq.SetFormDataFromValues(params)
	}

	result := &ListCertificatesResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
