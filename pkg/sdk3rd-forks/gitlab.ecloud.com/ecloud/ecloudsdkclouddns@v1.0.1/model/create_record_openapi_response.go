// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

type CreateRecordOpenapiResponseStateEnum string

// List of State
const (
	CreateRecordOpenapiResponseStateEnumError     CreateRecordOpenapiResponseStateEnum = "ERROR"
	CreateRecordOpenapiResponseStateEnumException CreateRecordOpenapiResponseStateEnum = "EXCEPTION"
	CreateRecordOpenapiResponseStateEnumForbidden CreateRecordOpenapiResponseStateEnum = "FORBIDDEN"
	CreateRecordOpenapiResponseStateEnumOk        CreateRecordOpenapiResponseStateEnum = "OK"
)

type CreateRecordOpenapiResponse struct {
	RequestId string `json:"requestId,omitempty"`

	ErrorMessage string `json:"errorMessage,omitempty"`

	ErrorCode string `json:"errorCode,omitempty"`

	State CreateRecordOpenapiResponseStateEnum `json:"state,omitempty"`

	Body *CreateRecordOpenapiResponseBody `json:"body,omitempty"`
}
