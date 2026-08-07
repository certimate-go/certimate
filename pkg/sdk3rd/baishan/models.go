package baishan

import (
	"encoding/json"
)

type Domain struct {
	Id         string `json:"id"`
	Domain     string `json:"domain"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	Cname      string `json:"cname"`
	Area       string `json:"area"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
}

type DomainCertificate struct {
	CertId         json.Number `json:"cert_id"`
	Name           string      `json:"name"`
	CertStartTime  string      `json:"cert_start_time"`
	CertExpireTime string      `json:"cert_expire_time"`
}

type DomainConfig struct {
	Https *DomainConfigHttps `json:"https"`
}

type DomainConfigHttps struct {
	CertId      json.Number `json:"cert_id"`
	ForceHttps  *string     `json:"force_https,omitempty"`
	EnableHttp2 *string     `json:"http2,omitempty"`
	EnableOcsp  *string     `json:"ocsp,omitempty"`
}
