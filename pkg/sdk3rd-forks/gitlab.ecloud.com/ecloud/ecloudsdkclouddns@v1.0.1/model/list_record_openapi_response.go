// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

type ListRecordOpenapiResponseStateEnum string

// List of State
const (
	ListRecordOpenapiResponseStateEnumError     ListRecordOpenapiResponseStateEnum = "ERROR"
	ListRecordOpenapiResponseStateEnumException ListRecordOpenapiResponseStateEnum = "EXCEPTION"
	ListRecordOpenapiResponseStateEnumForbidden ListRecordOpenapiResponseStateEnum = "FORBIDDEN"
	ListRecordOpenapiResponseStateEnumOk        ListRecordOpenapiResponseStateEnum = "OK"
)

type ListRecordOpenapiResponse struct {
	RequestId string `json:"requestId,omitempty"`

	ErrorMessage string `json:"errorMessage,omitempty"`

	ErrorCode string `json:"errorCode,omitempty"`

	State ListRecordOpenapiResponseStateEnum `json:"state,omitempty"`

	Body *ListRecordOpenapiResponseBody `json:"body,omitempty"`
}
