// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

type DeleteRecordOpenapiResponseStateEnum string

// List of State
const (
	DeleteRecordOpenapiResponseStateEnumError     DeleteRecordOpenapiResponseStateEnum = "ERROR"
	DeleteRecordOpenapiResponseStateEnumException DeleteRecordOpenapiResponseStateEnum = "EXCEPTION"
	DeleteRecordOpenapiResponseStateEnumForbidden DeleteRecordOpenapiResponseStateEnum = "FORBIDDEN"
	DeleteRecordOpenapiResponseStateEnumOk        DeleteRecordOpenapiResponseStateEnum = "OK"
)

type DeleteRecordOpenapiResponse struct {
	RequestId string `json:"requestId,omitempty"`

	ErrorMessage string `json:"errorMessage,omitempty"`

	ErrorCode string `json:"errorCode,omitempty"`

	State DeleteRecordOpenapiResponseStateEnum `json:"state,omitempty"`

	Body *[]DeleteRecordOpenapiResponseBody `json:"body,omitempty"`
}
