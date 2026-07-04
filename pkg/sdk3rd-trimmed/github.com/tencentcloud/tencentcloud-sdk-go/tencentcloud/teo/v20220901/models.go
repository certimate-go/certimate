package v20220901

import (
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	teo "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"
)

type (
	AccelerationDomain = teo.AccelerationDomain
	CertificateInfo    = teo.CertificateInfo
	ServerCertInfo     = teo.ServerCertInfo
)

type DescribeAccelerationDomainsRequest = teo.DescribeAccelerationDomainsRequest

type DescribeAccelerationDomainsResponse = teo.DescribeAccelerationDomainsResponse

// DescribeHostCertificatesRequest is the request structure
// for the DescribeHostCertificates API.
// Not exposed in Tencent Cloud official SDK or API document
// REF: https://docs.edgeone.site/#/?id=describehostcertificates
type DescribeHostCertificatesRequest struct {
	*tchttp.BaseRequest
	ZoneId  string                             `json:"ZoneId" name:"ZoneId"`
	Filters []*DescribeHostCertificatesFilters `json:"Filters,omitempty" name:"Filters"`
}

type DescribeHostCertificatesFilters struct {
	Name   string   `json:"Name" name:"Name"`
	Values []string `json:"Values" name:"Values"`
	Fuzzy  *bool    `json:"Fuzzy,omitempty" name:"Fuzzy"`
}

// DescribeHostCertificatesResponse is the response structure
// for the DescribeHostCertificates API.
// Not exposed in Tencent Cloud official SDK or API document
// REF: https://docs.edgeone.site/#/?id=describehostcertificates
type DescribeHostCertificatesResponse struct {
	*tchttp.BaseResponse
	Response DescribeHostCertificatesResponseParams `json:"Response"`
}

type DescribeHostCertificatesResponseParams struct {
	RequestId        string            `json:"RequestId"`
	HostCertificates []HostCertificate `json:"HostCertificates"`
	TotalCount       int               `json:"TotalCount"`
}

type HostCertificate struct {
	ApplyType        string             `json:"ApplyType"`
	ClientCertInfo   ClientCertInfo     `json:"ClientCertInfo"`
	Host             string             `json:"Host"`
	HostCertInfo     []HostCertInfoItem `json:"HostCertInfo"`
	Mode             string             `json:"Mode"`
	UpstreamCertInfo UpstreamCertInfo   `json:"UpstreamCertInfo"`
}

type ClientCertInfo struct {
	CertInfos []any  `json:"CertInfos"`
	Switch    string `json:"Switch"`
}

type HostCertInfoItem struct {
	Alias      string `json:"Alias"`
	CertId     string `json:"CertId"`
	DeployTime string `json:"DeployTime"`
	ExpireTime string `json:"ExpireTime"`
	SignAlgo   string `json:"SignAlgo"`
	Status     string `json:"Status"`
	Type       string `json:"Type"`
}

type UpstreamCertInfo struct {
	UpstreamCertificateVerify UpstreamCertificateVerify `json:"UpstreamCertificateVerify"`
	UpstreamMutualTLS         UpstreamMutualTLS         `json:"UpstreamMutualTLS"`
}

type UpstreamCertificateVerify struct {
	CustomCACerts    []interface{} `json:"CustomCACerts"` // 类型未知
	VerificationMode string        `json:"VerificationMode"`
}

type UpstreamMutualTLS struct {
	CertInfos []interface{} `json:"CertInfos"` // 类型未知
	Switch    string        `json:"Switch"`
}

type ModifyHostsCertificateRequest = teo.ModifyHostsCertificateRequest

type ModifyHostsCertificateResponse = teo.ModifyHostsCertificateResponse
