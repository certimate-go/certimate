// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

type DeleteRecordResponseStateEnum string

// List of State
const (
	DeleteRecordResponseStateEnumError     DeleteRecordResponseStateEnum = "ERROR"
	DeleteRecordResponseStateEnumException DeleteRecordResponseStateEnum = "EXCEPTION"
	DeleteRecordResponseStateEnumForbidden DeleteRecordResponseStateEnum = "FORBIDDEN"
	DeleteRecordResponseStateEnumOk        DeleteRecordResponseStateEnum = "OK"
)

type DeleteRecordResponse struct {
	RequestId string `json:"requestId,omitempty"`

	ErrorMessage string `json:"errorMessage,omitempty"`

	ErrorCode string `json:"errorCode,omitempty"`

	State DeleteRecordResponseStateEnum `json:"state,omitempty"`

	Body *[]DeleteRecordResponseBody `json:"body,omitempty"`
}
