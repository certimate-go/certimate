// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

type ModifyRecordResponseBodyTypeEnum string

// List of Type
const (
	ModifyRecordResponseBodyTypeEnumA      ModifyRecordResponseBodyTypeEnum = "A"
	ModifyRecordResponseBodyTypeEnumAaaa   ModifyRecordResponseBodyTypeEnum = "AAAA"
	ModifyRecordResponseBodyTypeEnumCaa    ModifyRecordResponseBodyTypeEnum = "CAA"
	ModifyRecordResponseBodyTypeEnumCmauth ModifyRecordResponseBodyTypeEnum = "CMAUTH"
	ModifyRecordResponseBodyTypeEnumCname  ModifyRecordResponseBodyTypeEnum = "CNAME"
	ModifyRecordResponseBodyTypeEnumMx     ModifyRecordResponseBodyTypeEnum = "MX"
	ModifyRecordResponseBodyTypeEnumNs     ModifyRecordResponseBodyTypeEnum = "NS"
	ModifyRecordResponseBodyTypeEnumPtr    ModifyRecordResponseBodyTypeEnum = "PTR"
	ModifyRecordResponseBodyTypeEnumRp     ModifyRecordResponseBodyTypeEnum = "RP"
	ModifyRecordResponseBodyTypeEnumSpf    ModifyRecordResponseBodyTypeEnum = "SPF"
	ModifyRecordResponseBodyTypeEnumSrv    ModifyRecordResponseBodyTypeEnum = "SRV"
	ModifyRecordResponseBodyTypeEnumTxt    ModifyRecordResponseBodyTypeEnum = "TXT"
	ModifyRecordResponseBodyTypeEnumUrl    ModifyRecordResponseBodyTypeEnum = "URL"
)

type ModifyRecordResponseBodyStateEnum string

// List of State
const (
	ModifyRecordResponseBodyStateEnumDisabled ModifyRecordResponseBodyStateEnum = "DISABLED"
	ModifyRecordResponseBodyStateEnumEnabled  ModifyRecordResponseBodyStateEnum = "ENABLED"
)

type ModifyRecordResponseBody struct {
	// 主机头
	Rr string `json:"rr,omitempty"`

	// 修改时间
	ModifiedTime string `json:"modifiedTime,omitempty"`

	// 线路中文名
	LineZh string `json:"lineZh,omitempty"`

	// 备注
	Description string `json:"description,omitempty"`

	// 线路ID
	LineId string `json:"lineId,omitempty"`

	// 权重值
	Weight *int32 `json:"weight,omitempty"`

	// MX优先级
	MxPri *int32 `json:"mxPri,omitempty"`

	// 记录类型
	Type ModifyRecordResponseBodyTypeEnum `json:"type,omitempty"`

	// 缓存的生命周期
	Ttl *int32 `json:"ttl,omitempty"`

	// 解析记录ID
	RecordId string `json:"recordId,omitempty"`

	// 域名名称
	DomainName string `json:"domainName,omitempty"`

	// 线路英文名
	LineEn string `json:"lineEn,omitempty"`

	// 状态
	State ModifyRecordResponseBodyStateEnum `json:"state,omitempty"`

	// 记录值
	Value string `json:"value,omitempty"`
}
