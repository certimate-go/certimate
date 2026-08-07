package btpanelgo

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
