package synologydsm

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type ImportCertificateRequest struct {
	ID          string `json:"id"         url:"id"`
	Description string `json:"desc"       url:"desc"`
	Key         string `json:"key"        url:"key"`
	Cert        string `json:"cert"       url:"cert"`
	InterCert   string `json:"inter_cert" url:"inter_cert"`
	AsDefault   bool   `json:"as_default" url:"as_default"`
}

type ImportCertificateResponse struct {
	sdkResponseBase
	Data *struct {
		RestartHttpd bool `json:"restart_httpd"`
	} `json:"data,omitempty"`
}

func (c *Client) ImportCertificate(req *ImportCertificateRequest) (*ImportCertificateResponse, error) {
	params := url.Values{
		"api":       {"SYNO.Core.Certificate"},
		"method":    {"import"},
		"version":   {"1"},
		"_sid":      {c.sid},
		"SynoToken": {c.synoToken},
	}

	httpreq, err := c.newRequest(http.MethodPost, fmt.Sprintf("/webapi/entry.cgi?%s", params.Encode()))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetMultipartField("key", "key.pem", "text/plain", strings.NewReader(req.Key))
		httpreq.SetMultipartField("cert", "cert.pem", "text/plain", strings.NewReader(req.Cert))
		httpreq.SetMultipartField("inter_cert", "chain.pem", "text/plain", strings.NewReader(req.InterCert))
		httpreq.SetMultipartField("id", "", "", strings.NewReader(req.ID))
		httpreq.SetMultipartField("desc", "", "", strings.NewReader(req.Description))
		if req.AsDefault {
			httpreq.SetMultipartField("as_default", "", "", strings.NewReader("true"))
		}
	}

	result := &ImportCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
