// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

type ModifyRecordResponseStateEnum string

// List of State
const (
	ModifyRecordResponseStateEnumError     ModifyRecordResponseStateEnum = "ERROR"
	ModifyRecordResponseStateEnumException ModifyRecordResponseStateEnum = "EXCEPTION"
	ModifyRecordResponseStateEnumForbidden ModifyRecordResponseStateEnum = "FORBIDDEN"
	ModifyRecordResponseStateEnumOk        ModifyRecordResponseStateEnum = "OK"
)

type ModifyRecordResponse struct {
	RequestId string `json:"requestId,omitempty"`

	ErrorMessage string `json:"errorMessage,omitempty"`

	ErrorCode string `json:"errorCode,omitempty"`

	State ModifyRecordResponseStateEnum `json:"state,omitempty"`

	Body *ModifyRecordResponseBody `json:"body,omitempty"`
}
