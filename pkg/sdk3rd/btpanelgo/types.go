package btpanel

import (
	"encoding/json"
)

type apiRequestWithBlob interface {
	GetBlob() []byte
}

type apiResponse interface {
	GetStatus() json.RawMessage
	GetMessage() *string
}

type apiResponseBase struct {
	Status  json.RawMessage `json:"status,omitempty"`
	Code    *int            `json:"code,omitempty"`
	Message *string         `json:"msg,omitempty"`
}

func (r *apiResponseBase) GetStatus() json.RawMessage {
	return r.Status
}

func (r *apiResponseBase) GetMessage() *string {
	return r.Message
}

type SiteData struct {
	Id            int32  `json:"id"`
	Name          string `json:"name"`
	ProjectType   string `json:"project_type"`
	ProjectConfig string `json:"project_config"`
	AddTime       string `json:"addTime"`
}

type PageData struct {
	Page    int32 `json:"page"`
	Limit   int32 `json:"limit"`
	Total   int32 `json:"total"`
	Start   int32 `json:"start"`
	End     int32 `json:"end"`
	MaxPage int32 `json:"maxPage"`
}
