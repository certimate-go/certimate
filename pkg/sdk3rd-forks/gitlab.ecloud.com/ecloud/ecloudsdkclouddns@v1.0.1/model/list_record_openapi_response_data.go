// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

type ListRecordOpenapiResponseDataTypeEnum string

// List of Type
const (
	ListRecordOpenapiResponseDataTypeEnumA      ListRecordOpenapiResponseDataTypeEnum = "A"
	ListRecordOpenapiResponseDataTypeEnumAaaa   ListRecordOpenapiResponseDataTypeEnum = "AAAA"
	ListRecordOpenapiResponseDataTypeEnumCname  ListRecordOpenapiResponseDataTypeEnum = "CNAME"
	ListRecordOpenapiResponseDataTypeEnumMx     ListRecordOpenapiResponseDataTypeEnum = "MX"
	ListRecordOpenapiResponseDataTypeEnumTxt    ListRecordOpenapiResponseDataTypeEnum = "TXT"
	ListRecordOpenapiResponseDataTypeEnumNs     ListRecordOpenapiResponseDataTypeEnum = "NS"
	ListRecordOpenapiResponseDataTypeEnumSpf    ListRecordOpenapiResponseDataTypeEnum = "SPF"
	ListRecordOpenapiResponseDataTypeEnumSrv    ListRecordOpenapiResponseDataTypeEnum = "SRV"
	ListRecordOpenapiResponseDataTypeEnumCaa    ListRecordOpenapiResponseDataTypeEnum = "CAA"
	ListRecordOpenapiResponseDataTypeEnumCmauth ListRecordOpenapiResponseDataTypeEnum = "CMAUTH"
)

type ListRecordOpenapiResponseDataTimedStatusEnum string

// List of TimedStatus
const (
	ListRecordOpenapiResponseDataTimedStatusEnumDisabled ListRecordOpenapiResponseDataTimedStatusEnum = "DISABLED"
	ListRecordOpenapiResponseDataTimedStatusEnumEnabled  ListRecordOpenapiResponseDataTimedStatusEnum = "ENABLED"
	ListRecordOpenapiResponseDataTimedStatusEnumTimed    ListRecordOpenapiResponseDataTimedStatusEnum = "TIMED"
)

type ListRecordOpenapiResponseDataStateEnum string

// List of State
const (
	ListRecordOpenapiResponseDataStateEnumDisabled ListRecordOpenapiResponseDataStateEnum = "DISABLED"
	ListRecordOpenapiResponseDataStateEnumEnabled  ListRecordOpenapiResponseDataStateEnum = "ENABLED"
)

type ListRecordOpenapiResponseData struct {
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
	Type ListRecordOpenapiResponseDataTypeEnum `json:"type,omitempty"`

	// 缓存的生命周期
	Ttl *int32 `json:"ttl,omitempty"`

	// 标签
	Tags *[]ListRecordOpenapiResponseTags `json:"tags,omitempty"`

	// 解析记录ID
	RecordId string `json:"recordId,omitempty"`

	// 定时状态
	TimedStatus ListRecordOpenapiResponseDataTimedStatusEnum `json:"timedStatus,omitempty"`

	// 域名名称
	DomainName string `json:"domainName,omitempty"`

	// 线路英文名
	LineEn string `json:"lineEn,omitempty"`

	// 状态
	State ListRecordOpenapiResponseDataStateEnum `json:"state,omitempty"`

	// 记录值
	Value string `json:"value,omitempty"`

	// 定时发布时间
	Pubdate string `json:"pubdate,omitempty"`
}
