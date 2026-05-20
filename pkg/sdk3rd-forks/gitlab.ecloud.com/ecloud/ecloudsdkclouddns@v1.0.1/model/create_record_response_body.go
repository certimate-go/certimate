// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

type CreateRecordResponseBodyTypeEnum string

// List of Type
const (
	CreateRecordResponseBodyTypeEnumA      CreateRecordResponseBodyTypeEnum = "A"
	CreateRecordResponseBodyTypeEnumAaaa   CreateRecordResponseBodyTypeEnum = "AAAA"
	CreateRecordResponseBodyTypeEnumCaa    CreateRecordResponseBodyTypeEnum = "CAA"
	CreateRecordResponseBodyTypeEnumCmauth CreateRecordResponseBodyTypeEnum = "CMAUTH"
	CreateRecordResponseBodyTypeEnumCname  CreateRecordResponseBodyTypeEnum = "CNAME"
	CreateRecordResponseBodyTypeEnumMx     CreateRecordResponseBodyTypeEnum = "MX"
	CreateRecordResponseBodyTypeEnumNs     CreateRecordResponseBodyTypeEnum = "NS"
	CreateRecordResponseBodyTypeEnumPtr    CreateRecordResponseBodyTypeEnum = "PTR"
	CreateRecordResponseBodyTypeEnumRp     CreateRecordResponseBodyTypeEnum = "RP"
	CreateRecordResponseBodyTypeEnumSpf    CreateRecordResponseBodyTypeEnum = "SPF"
	CreateRecordResponseBodyTypeEnumSrv    CreateRecordResponseBodyTypeEnum = "SRV"
	CreateRecordResponseBodyTypeEnumTxt    CreateRecordResponseBodyTypeEnum = "TXT"
	CreateRecordResponseBodyTypeEnumUrl    CreateRecordResponseBodyTypeEnum = "URL"
)

type CreateRecordResponseBodyTimedStatusEnum string

// List of TimedStatus
const (
	CreateRecordResponseBodyTimedStatusEnumDisabled CreateRecordResponseBodyTimedStatusEnum = "DISABLED"
	CreateRecordResponseBodyTimedStatusEnumEnabled  CreateRecordResponseBodyTimedStatusEnum = "ENABLED"
	CreateRecordResponseBodyTimedStatusEnumTimed    CreateRecordResponseBodyTimedStatusEnum = "TIMED"
)

type CreateRecordResponseBodyStateEnum string

// List of State
const (
	CreateRecordResponseBodyStateEnumDisabled CreateRecordResponseBodyStateEnum = "DISABLED"
	CreateRecordResponseBodyStateEnumEnabled  CreateRecordResponseBodyStateEnum = "ENABLED"
)

type CreateRecordResponseBody struct {
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
	Type CreateRecordResponseBodyTypeEnum `json:"type,omitempty"`

	// 缓存的生命周期
	Ttl *int32 `json:"ttl,omitempty"`

	// 标签
	Tags *[]CreateRecordResponseTags `json:"tags,omitempty"`

	// 解析记录ID
	RecordId string `json:"recordId,omitempty"`

	// 定时状态
	TimedStatus CreateRecordResponseBodyTimedStatusEnum `json:"timedStatus,omitempty"`

	// 域名名称
	DomainName string `json:"domainName,omitempty"`

	// 线路英文名
	LineEn string `json:"lineEn,omitempty"`

	// 状态
	State CreateRecordResponseBodyStateEnum `json:"state,omitempty"`

	// 记录值
	Value string `json:"value,omitempty"`

	// 定时发布时间
	Pubdate string `json:"pubdate,omitempty"`
}
