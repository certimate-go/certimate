package response

import (
	"encoding/json"

	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/response"
)

type ByteplusResponse struct {
	ResponseMetadata *ResponseMetadata
	Result           interface{}
}

func (r *ByteplusResponse) Unwrap() *response.ByteplusResponse {
	t := &response.ByteplusResponse{
		Result: r.Result,
	}
	if r.ResponseMetadata != nil {
		t.ResponseMetadata = r.ResponseMetadata.Unwrap()
	}
	return t
}

type ResponseMetadata struct {
	RequestId string
	Action    string
	Version   string
	Service   string
	Region    string
	HTTPCode  int
	Error     *Error
}

func (r *ResponseMetadata) Unwrap() *response.ResponseMetadata {
	t := &response.ResponseMetadata{
		RequestId: r.RequestId,
		Action:    r.Action,
		Version:   r.Version,
		Service:   r.Service,
		Region:    r.Region,
		HTTPCode:  r.HTTPCode,
	}
	if r.Error != nil {
		t.Error = r.Error.Unwrap()
	}
	return t
}

type Error struct {
	CodeN   int
	Code    string
	Message string
}

func (e *Error) UnmarshalJSON(data []byte) error {
	type errorAlias Error
	var decoded struct {
		*errorAlias
		Code json.RawMessage
	}
	decoded.errorAlias = (*errorAlias)(e)
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	if len(decoded.Code) == 0 || string(decoded.Code) == "null" {
		e.Code = ""
		return nil
	}
	if decoded.Code[0] == '"' {
		return json.Unmarshal(decoded.Code, &e.Code)
	}

	var code json.Number
	if err := json.Unmarshal(decoded.Code, &code); err != nil {
		return err
	}
	e.Code = code.String()
	return nil
}

func (e *Error) Unwrap() *response.Error {
	return &response.Error{
		CodeN:   e.CodeN,
		Code:    e.Code,
		Message: e.Message,
	}
}

type ByteplusSimpleError struct {
	HttpCode  int    `json:"HTTPCode"`
	ErrorCode string `json:"errorcode"`
	Message   string `json:"message"`
}

func (e *ByteplusSimpleError) Unwrap() *response.ByteplusSimpleError {
	return &response.ByteplusSimpleError{
		HttpCode:  e.HttpCode,
		ErrorCode: e.ErrorCode,
		Message:   e.Message,
	}
}
