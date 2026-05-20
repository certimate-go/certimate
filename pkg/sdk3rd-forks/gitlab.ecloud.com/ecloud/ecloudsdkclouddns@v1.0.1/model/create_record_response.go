// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

type CreateRecordResponseStateEnum string

// List of State
const (
	CreateRecordResponseStateEnumError     CreateRecordResponseStateEnum = "ERROR"
	CreateRecordResponseStateEnumException CreateRecordResponseStateEnum = "EXCEPTION"
	CreateRecordResponseStateEnumForbidden CreateRecordResponseStateEnum = "FORBIDDEN"
	CreateRecordResponseStateEnumOk        CreateRecordResponseStateEnum = "OK"
)

type CreateRecordResponse struct {
	RequestId string `json:"requestId,omitempty"`

	ErrorMessage string `json:"errorMessage,omitempty"`

	ErrorCode string `json:"errorCode,omitempty"`

	State CreateRecordResponseStateEnum `json:"state,omitempty"`

	Body *CreateRecordResponseBody `json:"body,omitempty"`
}
