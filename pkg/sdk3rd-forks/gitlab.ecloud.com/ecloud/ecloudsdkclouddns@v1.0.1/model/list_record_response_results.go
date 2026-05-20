// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

type ListRecordResponseResultsTypeEnum string

// List of Type
const (
	ListRecordResponseResultsTypeEnumA      ListRecordResponseResultsTypeEnum = "A"
	ListRecordResponseResultsTypeEnumAaaa   ListRecordResponseResultsTypeEnum = "AAAA"
	ListRecordResponseResultsTypeEnumCaa    ListRecordResponseResultsTypeEnum = "CAA"
	ListRecordResponseResultsTypeEnumCmauth ListRecordResponseResultsTypeEnum = "CMAUTH"
	ListRecordResponseResultsTypeEnumCname  ListRecordResponseResultsTypeEnum = "CNAME"
	ListRecordResponseResultsTypeEnumMx     ListRecordResponseResultsTypeEnum = "MX"
	ListRecordResponseResultsTypeEnumNs     ListRecordResponseResultsTypeEnum = "NS"
	ListRecordResponseResultsTypeEnumPtr    ListRecordResponseResultsTypeEnum = "PTR"
	ListRecordResponseResultsTypeEnumRp     ListRecordResponseResultsTypeEnum = "RP"
	ListRecordResponseResultsTypeEnumSpf    ListRecordResponseResultsTypeEnum = "SPF"
	ListRecordResponseResultsTypeEnumSrv    ListRecordResponseResultsTypeEnum = "SRV"
	ListRecordResponseResultsTypeEnumTxt    ListRecordResponseResultsTypeEnum = "TXT"
	ListRecordResponseResultsTypeEnumUrl    ListRecordResponseResultsTypeEnum = "URL"
)

type ListRecordResponseResultsTimedStatusEnum string

// List of TimedStatus
const (
	ListRecordResponseResultsTimedStatusEnumDisabled ListRecordResponseResultsTimedStatusEnum = "DISABLED"
	ListRecordResponseResultsTimedStatusEnumEnabled  ListRecordResponseResultsTimedStatusEnum = "ENABLED"
	ListRecordResponseResultsTimedStatusEnumTimed    ListRecordResponseResultsTimedStatusEnum = "TIMED"
)

type ListRecordResponseResultsStateEnum string

// List of State
const (
	ListRecordResponseResultsStateEnumDisabled ListRecordResponseResultsStateEnum = "DISABLED"
	ListRecordResponseResultsStateEnumEnabled  ListRecordResponseResultsStateEnum = "ENABLED"
)

type ListRecordResponseResults struct {
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
	Type ListRecordResponseResultsTypeEnum `json:"type,omitempty"`

	// 缓存的生命周期
	Ttl *int32 `json:"ttl,omitempty"`

	// 解析记录ID
	RecordId string `json:"recordId,omitempty"`

	// 定时状态
	TimedStatus ListRecordResponseResultsTimedStatusEnum `json:"timedStatus,omitempty"`

	// 域名名称
	DomainName string `json:"domainName,omitempty"`

	// 线路英文名
	LineEn string `json:"lineEn,omitempty"`

	// 状态
	State ListRecordResponseResultsStateEnum `json:"state,omitempty"`

	// 记录值
	Value string `json:"value,omitempty"`

	// 定时发布时间
	Pubdate string `json:"pubdate,omitempty"`
}
