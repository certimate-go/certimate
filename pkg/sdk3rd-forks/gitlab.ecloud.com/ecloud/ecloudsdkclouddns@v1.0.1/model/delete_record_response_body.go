// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

type DeleteRecordResponseBodyCodeEnum string

// List of Code
const (
	DeleteRecordResponseBodyCodeEnumError   DeleteRecordResponseBodyCodeEnum = "ERROR"
	DeleteRecordResponseBodyCodeEnumSuccess DeleteRecordResponseBodyCodeEnum = "SUCCESS"
)

type DeleteRecordResponseBody struct {
	// 结果说明
	Msg string `json:"msg,omitempty"`

	// 解析记录ID
	RecordId string `json:"recordId,omitempty"`

	// 结果码
	Code DeleteRecordResponseBodyCodeEnum `json:"code,omitempty"`

	// 域名
	DomainName string `json:"domainName,omitempty"`
}
