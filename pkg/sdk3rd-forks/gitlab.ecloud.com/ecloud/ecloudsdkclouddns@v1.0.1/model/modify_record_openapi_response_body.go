// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

type ModifyRecordOpenapiResponseBodyTypeEnum string

// List of Type
const (
	ModifyRecordOpenapiResponseBodyTypeEnumA      ModifyRecordOpenapiResponseBodyTypeEnum = "A"
	ModifyRecordOpenapiResponseBodyTypeEnumAaaa   ModifyRecordOpenapiResponseBodyTypeEnum = "AAAA"
	ModifyRecordOpenapiResponseBodyTypeEnumCname  ModifyRecordOpenapiResponseBodyTypeEnum = "CNAME"
	ModifyRecordOpenapiResponseBodyTypeEnumMx     ModifyRecordOpenapiResponseBodyTypeEnum = "MX"
	ModifyRecordOpenapiResponseBodyTypeEnumTxt    ModifyRecordOpenapiResponseBodyTypeEnum = "TXT"
	ModifyRecordOpenapiResponseBodyTypeEnumNs     ModifyRecordOpenapiResponseBodyTypeEnum = "NS"
	ModifyRecordOpenapiResponseBodyTypeEnumSpf    ModifyRecordOpenapiResponseBodyTypeEnum = "SPF"
	ModifyRecordOpenapiResponseBodyTypeEnumSrv    ModifyRecordOpenapiResponseBodyTypeEnum = "SRV"
	ModifyRecordOpenapiResponseBodyTypeEnumCaa    ModifyRecordOpenapiResponseBodyTypeEnum = "CAA"
	ModifyRecordOpenapiResponseBodyTypeEnumCmauth ModifyRecordOpenapiResponseBodyTypeEnum = "CMAUTH"
)

type ModifyRecordOpenapiResponseBodyTimedStatusEnum string

// List of TimedStatus
const (
	ModifyRecordOpenapiResponseBodyTimedStatusEnumDisabled ModifyRecordOpenapiResponseBodyTimedStatusEnum = "DISABLED"
	ModifyRecordOpenapiResponseBodyTimedStatusEnumEnabled  ModifyRecordOpenapiResponseBodyTimedStatusEnum = "ENABLED"
	ModifyRecordOpenapiResponseBodyTimedStatusEnumTimed    ModifyRecordOpenapiResponseBodyTimedStatusEnum = "TIMED"
)

type ModifyRecordOpenapiResponseBodyStateEnum string

// List of State
const (
	ModifyRecordOpenapiResponseBodyStateEnumDisabled ModifyRecordOpenapiResponseBodyStateEnum = "DISABLED"
	ModifyRecordOpenapiResponseBodyStateEnumEnabled  ModifyRecordOpenapiResponseBodyStateEnum = "ENABLED"
)

type ModifyRecordOpenapiResponseBody struct {
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
	Type ModifyRecordOpenapiResponseBodyTypeEnum `json:"type,omitempty"`

	// 缓存的生命周期
	Ttl *int32 `json:"ttl,omitempty"`

	// 标签
	Tags *[]ModifyRecordOpenapiResponseTags `json:"tags,omitempty"`

	// 解析记录ID
	RecordId string `json:"recordId,omitempty"`

	// 定时状态
	TimedStatus ModifyRecordOpenapiResponseBodyTimedStatusEnum `json:"timedStatus,omitempty"`

	// 域名名称
	DomainName string `json:"domainName,omitempty"`

	// 线路英文名
	LineEn string `json:"lineEn,omitempty"`

	// 状态
	State ModifyRecordOpenapiResponseBodyStateEnum `json:"state,omitempty"`

	// 记录值
	Value string `json:"value,omitempty"`

	// 定时发布时间
	Pubdate string `json:"pubdate,omitempty"`
}
