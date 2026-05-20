// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

type CreateRecordOpenapiResponseBodyTypeEnum string

// List of Type
const (
	CreateRecordOpenapiResponseBodyTypeEnumA      CreateRecordOpenapiResponseBodyTypeEnum = "A"
	CreateRecordOpenapiResponseBodyTypeEnumAaaa   CreateRecordOpenapiResponseBodyTypeEnum = "AAAA"
	CreateRecordOpenapiResponseBodyTypeEnumCname  CreateRecordOpenapiResponseBodyTypeEnum = "CNAME"
	CreateRecordOpenapiResponseBodyTypeEnumMx     CreateRecordOpenapiResponseBodyTypeEnum = "MX"
	CreateRecordOpenapiResponseBodyTypeEnumTxt    CreateRecordOpenapiResponseBodyTypeEnum = "TXT"
	CreateRecordOpenapiResponseBodyTypeEnumNs     CreateRecordOpenapiResponseBodyTypeEnum = "NS"
	CreateRecordOpenapiResponseBodyTypeEnumSpf    CreateRecordOpenapiResponseBodyTypeEnum = "SPF"
	CreateRecordOpenapiResponseBodyTypeEnumSrv    CreateRecordOpenapiResponseBodyTypeEnum = "SRV"
	CreateRecordOpenapiResponseBodyTypeEnumCaa    CreateRecordOpenapiResponseBodyTypeEnum = "CAA"
	CreateRecordOpenapiResponseBodyTypeEnumCmauth CreateRecordOpenapiResponseBodyTypeEnum = "CMAUTH"
)

type CreateRecordOpenapiResponseBodyStateEnum string

// List of State
const (
	CreateRecordOpenapiResponseBodyStateEnumDisabled CreateRecordOpenapiResponseBodyStateEnum = "DISABLED"
	CreateRecordOpenapiResponseBodyStateEnumEnabled  CreateRecordOpenapiResponseBodyStateEnum = "ENABLED"
)

type CreateRecordOpenapiResponseBody struct {
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
	Type CreateRecordOpenapiResponseBodyTypeEnum `json:"type,omitempty"`

	// 缓存的生命周期
	Ttl *int32 `json:"ttl,omitempty"`

	// 标签
	Tags *[]CreateRecordOpenapiResponseTags `json:"tags,omitempty"`

	// 解析记录ID
	RecordId string `json:"recordId,omitempty"`

	// 域名名称
	DomainName string `json:"domainName,omitempty"`

	// 线路英文名
	LineEn string `json:"lineEn,omitempty"`

	// 状态
	State CreateRecordOpenapiResponseBodyStateEnum `json:"state,omitempty"`

	// 记录值
	Value string `json:"value,omitempty"`

	// 定时发布时间
	Pubdate string `json:"pubdate,omitempty"`
}
