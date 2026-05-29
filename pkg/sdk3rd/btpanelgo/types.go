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

type SiteData struct {
	Id          int32  `json:"id"`
	ProjectType string `json:"project_type"`
	Name        string `json:"name"`
	Note        string `json:"ps"`
	Status      string `json:"status"`
	SSLInfo     []*struct {
		Name   string `json:"name"`
		Status bool   `json:"status"`
	} `json:"ssl_info"`
	AddTime string `json:"addtime"`
}

type PageData struct {
	Page    int32 `json:"page"`
	Limit   int32 `json:"limit"`
	Total   int32 `json:"total"`
	Start   int32 `json:"start"`
	End     int32 `json:"end"`
	MaxPage int32 `json:"maxPage"`
}
