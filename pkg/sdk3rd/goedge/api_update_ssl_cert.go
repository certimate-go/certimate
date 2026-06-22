package goedge

import (
	"context"
	"net/http"
)

type UpdateSSLCertRequest struct {
	SSLCertId   int64    `json:"sslCertId"`
	IsOn        bool     `json:"isOn"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ServerName  string   `json:"serverName"`
	IsCA        bool     `json:"isCA"`
	CertData    string   `json:"certData"`
	KeyData     string   `json:"keyData"`
	TimeBeginAt int64    `json:"timeBeginAt"`
	TimeEndAt   int64    `json:"timeEndAt"`
	DNSNames    []string `json:"dnsNames"`
	CommonNames []string `json:"commonNames"`
}

type UpdateSSLCertResponse struct {
	sdkResponseBase
}

func (c *Client) UpdateSSLCert(req *UpdateSSLCertRequest) (*UpdateSSLCertResponse, error) {
	return c.UpdateSSLCertWithContext(context.Background(), req)
}

func (c *Client) UpdateSSLCertWithContext(ctx context.Context, req *UpdateSSLCertRequest) (*UpdateSSLCertResponse, error) {
	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodPost, "/SSLCertService/updateSSLCert")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &UpdateSSLCertResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
