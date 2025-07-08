package qiniu

import (
	"context"
	"net/http"
	"net/url"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/client"
)

type SslCertManager struct {
	client *client.Client
}

func NewSslCertManager(mac *auth.Credentials) *SslCertManager {
	if mac == nil {
		mac = auth.Default()
	}

	client := &client.Client{Client: &http.Client{Transport: newTransport(mac, nil)}}
	return &SslCertManager{client: client}
}

type GetSslCertListResponse struct {
	Code  *int    `json:"code,omitempty"`
	Error *string `json:"error,omitempty"`
	Certs []*struct {
		CertID           string   `json:"certid"`
		Name             string   `json:"name"`
		CommonName       string   `json:"common_name"`
		DnsNames         []string `json:"dnsnames"`
		CreateTime       int64    `json:"create_time"`
		NotBefore        int64    `json:"not_before"`
		NotAfter         int64    `json:"not_after"`
		ProductType      string   `json:"product_type"`
		ProductShortName string   `json:"product_short_name,omitempty"`
		OrderId          string   `json:"orderid,omitempty"`
		CertType         string   `json:"cert_type"`
		Encrypt          string   `json:"encrypt"`
		EncryptParameter string   `json:"encryptParameter,omitempty"`
		Enable           bool     `json:"enable"`
	} `json:"certs"`
	Marker string `json:"marker"`
}

func (m *SslCertManager) GetSslCertList(ctx context.Context, marker string, limit int32) (*GetSslCertListResponse, error) {
	resp := new(GetSslCertListResponse)
	if err := m.client.Call(ctx, resp, http.MethodGet, urlf("sslcert?marker=%s&limit=%d", url.QueryEscape(marker), limit), nil); err != nil {
		return nil, err
	}
	return resp, nil
}

type UploadSslCertRequest struct {
	Name        string `json:"name"`
	CommonName  string `json:"common_name"`
	Certificate string `json:"ca"`
	PrivateKey  string `json:"pri"`
}

type UploadSslCertResponse struct {
	Code   *int    `json:"code,omitempty"`
	Error  *string `json:"error,omitempty"`
	CertID string  `json:"certID"`
}

func (m *SslCertManager) UploadSslCert(ctx context.Context, name string, commonName string, certificate string, privateKey string) (*UploadSslCertResponse, error) {
	req := &UploadSslCertRequest{
		Name:        name,
		CommonName:  commonName,
		Certificate: certificate,
		PrivateKey:  privateKey,
	}
	resp := new(UploadSslCertResponse)
	if err := m.client.CallWithJson(ctx, resp, http.MethodPost, urlf("sslcert"), nil, req); err != nil {
		return nil, err
	}
	return resp, nil
}
