package elb

import (
	"bytes"
	"encoding/json"
	"strconv"
)

type sdkResponse interface {
	GetStatusCode() string
	GetMessage() string
	GetError() string
	GetDescription() string
}

type sdkResponseBase struct {
	StatusCode  json.RawMessage `json:"statusCode,omitempty"`
	Message     *string         `json:"message,omitempty"`
	Error       *string         `json:"error,omitempty"`
	Description *string         `json:"description,omitempty"`
	RequestId   *string         `json:"requestId,omitempty"`
}

func (r *sdkResponseBase) GetStatusCode() string {
	if r.StatusCode == nil {
		return ""
	}

	decoder := json.NewDecoder(bytes.NewReader(r.StatusCode))
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

func (r *sdkResponseBase) GetDescription() string {
	if r.Description == nil {
		return ""
	}

	return *r.Description
}

var _ sdkResponse = (*sdkResponseBase)(nil)
