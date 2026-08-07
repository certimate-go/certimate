package lvdn

import (
	"bytes"
	"encoding/json"
	"strconv"
)

type sdkResponse interface {
	GetCode() string
	GetMessage() string
	GetError() string
	GetErrorMessage() string
}

type sdkResponseBase struct {
	Code         json.RawMessage `json:"code,omitempty"`
	Message      *string         `json:"message,omitempty"`
	Error        *string         `json:"error,omitempty"`
	ErrorMessage *string         `json:"errorMessage,omitempty"`
	RequestId    *string         `json:"requestId,omitempty"`
}

func (r *sdkResponseBase) GetCode() string {
	if r.Code == nil {
		return ""
	}

	decoder := json.NewDecoder(bytes.NewReader(r.Code))
	token, err := decoder.Token()
	if err != nil {
		return ""
	}

	switch t := token.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case json.Number:
		return t.String()
	default:
		return ""
	}
}

func (r *sdkResponseBase) GetMessage() string {
	if r.Message == nil {
		return ""
	}

	return *r.Message
}

func (r *sdkResponseBase) GetError() string {
	if r.Error == nil {
		return ""
	}

	return *r.Error
}

func (r *sdkResponseBase) GetErrorMessage() string {
	if r.ErrorMessage == nil {
		return ""
	}

	return *r.ErrorMessage
}

var _ sdkResponse = (*sdkResponseBase)(nil)
