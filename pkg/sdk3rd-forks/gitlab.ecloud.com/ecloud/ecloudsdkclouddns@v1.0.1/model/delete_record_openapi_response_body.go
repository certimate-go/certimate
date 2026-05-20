// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

type DeleteRecordOpenapiResponseBodyCodeEnum string

// List of Code
const (
	DeleteRecordOpenapiResponseBodyCodeEnumError   DeleteRecordOpenapiResponseBodyCodeEnum = "ERROR"
	DeleteRecordOpenapiResponseBodyCodeEnumSuccess DeleteRecordOpenapiResponseBodyCodeEnum = "SUCCESS"
)

type DeleteRecordOpenapiResponseBody struct {
	// 结果说明
	Msg string `json:"msg,omitempty"`

	// 解析记录ID
	RecordId string `json:"recordId,omitempty"`

	// 结果码
	Code DeleteRecordOpenapiResponseBodyCodeEnum `json:"code,omitempty"`

	// 域名
	DomainName string `json:"domainName,omitempty"`
}
