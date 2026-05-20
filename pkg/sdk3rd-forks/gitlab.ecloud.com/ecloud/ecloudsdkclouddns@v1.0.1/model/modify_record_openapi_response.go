// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

type ModifyRecordOpenapiResponseStateEnum string

// List of State
const (
	ModifyRecordOpenapiResponseStateEnumError     ModifyRecordOpenapiResponseStateEnum = "ERROR"
	ModifyRecordOpenapiResponseStateEnumException ModifyRecordOpenapiResponseStateEnum = "EXCEPTION"
	ModifyRecordOpenapiResponseStateEnumForbidden ModifyRecordOpenapiResponseStateEnum = "FORBIDDEN"
	ModifyRecordOpenapiResponseStateEnumOk        ModifyRecordOpenapiResponseStateEnum = "OK"
)

type ModifyRecordOpenapiResponse struct {
	RequestId string `json:"requestId,omitempty"`

	ErrorMessage string `json:"errorMessage,omitempty"`

	ErrorCode string `json:"errorCode,omitempty"`

	State ModifyRecordOpenapiResponseStateEnum `json:"state,omitempty"`

	Body *ModifyRecordOpenapiResponseBody `json:"body,omitempty"`
}
