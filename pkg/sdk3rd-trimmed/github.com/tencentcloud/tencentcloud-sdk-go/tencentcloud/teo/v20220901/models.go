package v20220901

import (
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	teo "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"
)

type (
	AccelerationDomain = teo.AccelerationDomain
	AdvancedFilter     = teo.AdvancedFilter
	CertificateInfo    = teo.CertificateInfo
	MutualTLS          = teo.MutualTLS
	ServerCertInfo     = teo.ServerCertInfo
	UpstreamCertInfo   = teo.UpstreamCertInfo
)

type DescribeAccelerationDomainsRequest = teo.DescribeAccelerationDomainsRequest

type DescribeAccelerationDomainsResponse = teo.DescribeAccelerationDomainsResponse

type DescribeHostCertificatesRequest struct {
	*tchttp.BaseRequest
	ZoneId  *string           `json:"ZoneId,omitnil,omitempty" name:"ZoneId"`
	Filters []*AdvancedFilter `json:"Filters,omitnil,omitempty" name:"Filters"`
}

type DescribeHostCertificatesResponse struct {
	*tchttp.BaseResponse
	Response *DescribeHostCertificatesResponseParams `json:"Response" name:"Response"`
}

type DescribeHostCertificatesResponseParams struct {
	RequestId        *string            `json:"RequestId,omitnil,omitempty"	name:"RequestId"`
	TotalCount       *int64             `json:"TotalCount,omitnil,omitempty" 	name:"TotalCount"`
	HostCertificates []*HostCertificate `json:"HostCertificates,omitnil,omitempty" 	name:"HostCertificates"`
}

type ModifyHostsCertificateRequest = teo.ModifyHostsCertificateRequest

type ModifyHostsCertificateResponse = teo.ModifyHostsCertificateResponse

type HostCertificate struct {
	ApplyType        *string            `json:"ApplyType,omitnil,omitempty" name:"ApplyType"`
	ClientCertInfo   *MutualTLS         `json:"ClientCertInfo,omitnil,omitempty" name:"ClientCertInfo"`
	Host             *string            `json:"Host,omitnil,omitempty" name:"Host"`
	HostCertInfo     []*CertificateInfo `json:"HostCertInfo,omitnil,omitempty" name:"HostCertInfo"`
	Mode             *string            `json:"Mode,omitnil,omitempty" name:"Mode"`
	ServerCertInfo   []*ServerCertInfo  `json:"ServerCertInfo,omitnil,omitempty" name:"ServerCertInfo"`
	UpstreamCertInfo *UpstreamCertInfo  `json:"UpstreamCertInfo,omitnil,omitempty" name:"UpstreamCertInfo"`
}
