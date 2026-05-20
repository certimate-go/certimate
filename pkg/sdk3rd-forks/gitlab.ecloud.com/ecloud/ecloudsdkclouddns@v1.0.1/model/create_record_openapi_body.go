// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

import (
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/position"
)

type CreateRecordOpenapiBodyTypeEnum string

// List of Type
const (
	CreateRecordOpenapiBodyTypeEnumA      CreateRecordOpenapiBodyTypeEnum = "A"
	CreateRecordOpenapiBodyTypeEnumAaaa   CreateRecordOpenapiBodyTypeEnum = "AAAA"
	CreateRecordOpenapiBodyTypeEnumCname  CreateRecordOpenapiBodyTypeEnum = "CNAME"
	CreateRecordOpenapiBodyTypeEnumMx     CreateRecordOpenapiBodyTypeEnum = "MX"
	CreateRecordOpenapiBodyTypeEnumTxt    CreateRecordOpenapiBodyTypeEnum = "TXT"
	CreateRecordOpenapiBodyTypeEnumNs     CreateRecordOpenapiBodyTypeEnum = "NS"
	CreateRecordOpenapiBodyTypeEnumSpf    CreateRecordOpenapiBodyTypeEnum = "SPF"
	CreateRecordOpenapiBodyTypeEnumSrv    CreateRecordOpenapiBodyTypeEnum = "SRV"
	CreateRecordOpenapiBodyTypeEnumCaa    CreateRecordOpenapiBodyTypeEnum = "CAA"
	CreateRecordOpenapiBodyTypeEnumCmauth CreateRecordOpenapiBodyTypeEnum = "CMAUTH"
)

type CreateRecordOpenapiBody struct {
	position.Body
	// 主机头
	Rr string `json:"rr"`

	// 域名名称
	DomainName string `json:"domainName"`

	// 备注
	Description string `json:"description,omitempty"`

	// 线路ID
	LineId string `json:"lineId"`

	// MX优先级，若“记录类型”选择”MX”，则需要配置该参数,默认是5
	MxPri *int32 `json:"mxPri,omitempty"`

	// 记录类型
	Type CreateRecordOpenapiBodyTypeEnum `json:"type"`

	// 缓存的生命周期，默认可配置600s
	Ttl *int32 `json:"ttl,omitempty"`

	// 记录值
	Value string `json:"value"`
}
