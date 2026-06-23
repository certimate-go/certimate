package btpanelgo

import (
	"encoding/json"
)

type sdkResponse interface {
	GetStatus() json.RawMessage
	GetMessage() *string
}

type sdkResponseBase struct {
	Status  json.RawMessage `json:"status,omitempty"`
	Code    *int            `json:"code,omitempty"`
	Message *string         `json:"msg,omitempty"`
}

func (r *sdkResponseBase) GetStatus() json.RawMessage {
	return r.Status
}

func (r *sdkResponseBase) GetMessage() *string {
	return r.Message
}

var _ sdkResponse = (*sdkResponseBase)(nil)
