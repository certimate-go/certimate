// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

import (
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/position"
)

type ModifyRecordOpenapiBodyTypeEnum string

// List of Type
const (
	ModifyRecordOpenapiBodyTypeEnumA      ModifyRecordOpenapiBodyTypeEnum = "A"
	ModifyRecordOpenapiBodyTypeEnumAaaa   ModifyRecordOpenapiBodyTypeEnum = "AAAA"
	ModifyRecordOpenapiBodyTypeEnumCname  ModifyRecordOpenapiBodyTypeEnum = "CNAME"
	ModifyRecordOpenapiBodyTypeEnumMx     ModifyRecordOpenapiBodyTypeEnum = "MX"
	ModifyRecordOpenapiBodyTypeEnumTxt    ModifyRecordOpenapiBodyTypeEnum = "TXT"
	ModifyRecordOpenapiBodyTypeEnumNs     ModifyRecordOpenapiBodyTypeEnum = "NS"
	ModifyRecordOpenapiBodyTypeEnumSpf    ModifyRecordOpenapiBodyTypeEnum = "SPF"
	ModifyRecordOpenapiBodyTypeEnumSrv    ModifyRecordOpenapiBodyTypeEnum = "SRV"
	ModifyRecordOpenapiBodyTypeEnumCaa    ModifyRecordOpenapiBodyTypeEnum = "CAA"
	ModifyRecordOpenapiBodyTypeEnumCmauth ModifyRecordOpenapiBodyTypeEnum = "CMAUTH"
)

type ModifyRecordOpenapiBody struct {
	position.Body
	// 解析记录ID
	RecordId string `json:"recordId"`

	// 主机头
	Rr string `json:"rr,omitempty"`

	// 域名名称
	DomainName string `json:"domainName"`

	// 备注
	Description string `json:"description,omitempty"`

	// 线路ID
	LineId string `json:"lineId,omitempty"`

	// MX优先级
	MxPri *int32 `json:"mxPri,omitempty"`

	// 记录类型
	Type ModifyRecordOpenapiBodyTypeEnum `json:"type,omitempty"`

	// 缓存的生命周期
	Ttl *int32 `json:"ttl,omitempty"`

	// 记录值
	Value string `json:"value,omitempty"`
}
