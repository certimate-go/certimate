package qiniu

import (
	"context"
	"net/http"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/client"
)

type CdnManager struct {
	client *client.Client
}

func NewCdnManager(mac *auth.Credentials) *CdnManager {
	if mac == nil {
		mac = auth.Default()
	}

	client := &client.Client{Client: &http.Client{Transport: newTransport(mac, nil)}}
	return &CdnManager{client: client}
}

type GetDomainInfoResponse struct {
	Code  *int    `json:"code,omitempty"`
	Error *string `json:"error,omitempty"`
	Name  string  `json:"name"`
	Type  string  `json:"type"`
	CName string  `json:"cname"`
	Https *struct {
		CertID      string `json:"certId"`
		ForceHttps  bool   `json:"forceHttps"`
		Http2Enable bool   `json:"http2Enable"`
	} `json:"https"`
	PareDomain         string `json:"pareDomain"`
	OperationType      string `json:"operationType"`
	OperatingState     string `json:"operatingState"`
	OperatingStateDesc string `json:"operatingStateDesc"`
	CreateAt           string `json:"createAt"`
	ModifyAt           string `json:"modifyAt"`
}

func (m *CdnManager) GetDomainInfo(ctx context.Context, domain string) (*GetDomainInfoResponse, error) {
	resp := new(GetDomainInfoResponse)
	if err := m.client.Call(ctx, resp, http.MethodGet, urlf("domain/%s", domain), nil); err != nil {
		return nil, err
	}
	return resp, nil
}

type ModifyDomainHttpsConfRequest struct {
	CertID      string `json:"certId"`
	ForceHttps  bool   `json:"forceHttps"`
	Http2Enable bool   `json:"http2Enable"`
}

type ModifyDomainHttpsConfResponse struct {
	Code  *int    `json:"code,omitempty"`
	Error *string `json:"error,omitempty"`
}

func (m *CdnManager) ModifyDomainHttpsConf(ctx context.Context, domain string, certId string, forceHttps bool, http2Enable bool) (*ModifyDomainHttpsConfResponse, error) {
	req := &ModifyDomainHttpsConfRequest{
		CertID:      certId,
		ForceHttps:  forceHttps,
		Http2Enable: http2Enable,
	}
	resp := new(ModifyDomainHttpsConfResponse)
	if err := m.client.CallWithJson(ctx, resp, http.MethodPut, urlf("domain/%s/httpsconf", domain), nil, req); err != nil {
		return nil, err
	}
	return resp, nil
}

type EnableDomainHttpsRequest struct {
	CertID      string `json:"certId"`
	ForceHttps  bool   `json:"forceHttps"`
	Http2Enable bool   `json:"http2Enable"`
}

type EnableDomainHttpsResponse struct {
	Code  *int    `json:"code,omitempty"`
	Error *string `json:"error,omitempty"`
}

func (m *CdnManager) EnableDomainHttps(ctx context.Context, domain string, certId string, forceHttps bool, http2Enable bool) (*EnableDomainHttpsResponse, error) {
	req := &EnableDomainHttpsRequest{
		CertID:      certId,
		ForceHttps:  forceHttps,
		Http2Enable: http2Enable,
	}
	resp := new(EnableDomainHttpsResponse)
	if err := m.client.CallWithJson(ctx, resp, http.MethodPut, urlf("domain/%s/sslize", domain), nil, req); err != nil {
		return nil, err
	}
	return resp, nil
}
