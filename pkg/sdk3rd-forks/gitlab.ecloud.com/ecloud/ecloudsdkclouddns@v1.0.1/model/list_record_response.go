// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

type ListRecordResponseStateEnum string

// List of State
const (
	ListRecordResponseStateEnumError     ListRecordResponseStateEnum = "ERROR"
	ListRecordResponseStateEnumException ListRecordResponseStateEnum = "EXCEPTION"
	ListRecordResponseStateEnumForbidden ListRecordResponseStateEnum = "FORBIDDEN"
	ListRecordResponseStateEnumOk        ListRecordResponseStateEnum = "OK"
)

type ListRecordResponse struct {
	RequestId string `json:"requestId,omitempty"`

	ErrorMessage string `json:"errorMessage,omitempty"`

	ErrorCode string `json:"errorCode,omitempty"`

	State ListRecordResponseStateEnum `json:"state,omitempty"`

	Body *ListRecordResponseBody `json:"body,omitempty"`
}
