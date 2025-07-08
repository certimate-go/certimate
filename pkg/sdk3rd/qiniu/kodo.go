package qiniu

import (
	"context"
	"net/http"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/client"
)

type KodoManager struct {
	client *client.Client
}

func NewKodoManager(mac *auth.Credentials) *KodoManager {
	if mac == nil {
		mac = auth.Default()
	}

	client := &client.Client{Client: &http.Client{Transport: newTransport(mac, nil)}}
	return &KodoManager{client: client}
}

type BindBucketCertRequest struct {
	CertID string `json:"certid"`
	Domain string `json:"domain"`
}

type BindBucketCertResponse struct {
	Code  *int    `json:"code,omitempty"`
	Error *string `json:"error,omitempty"`
}

func (m *KodoManager) BindBucketCert(ctx context.Context, domain string, certId string) (*BindBucketCertResponse, error) {
	req := &BindBucketCertRequest{
		CertID: certId,
		Domain: domain,
	}
	resp := new(BindBucketCertResponse)
	if err := m.client.CallWithJson(ctx, resp, http.MethodPut, urlf("cert/bind"), nil, req); err != nil {
		return nil, err
	}
	return resp, nil
}
