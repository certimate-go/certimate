package btwaf

type apiResponse interface {
	GetCode() int32
}

type apiResponseBase struct {
	Code *int32 `json:"code,omitempty"`
}

func (r *apiResponseBase) GetCode() int32 {
	if r.Code == nil {
		return 0
	}

	return *r.Code
}

var _ apiResponse = (*apiResponseBase)(nil)

type SiteRecord struct {
	SiteId     string `json:"site_id"`
	SiteName   string `json:"site_name"`
	Type       string `json:"types"`
	Status     int32  `json:"status"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
}

type SiteServerInfo struct {
	ListenSSLPorts *[]int32           `json:"listen_ssl_port,omitempty"`
	SSL            *SiteServerSSLInfo `json:"ssl,omitempty"`
}

type SiteServerInfoMod struct {
	ListenSSLPorts *[]string          `json:"listen_ssl_port,omitempty"`
	SSL            *SiteServerSSLInfo `json:"ssl,omitempty"`
}

type SiteServerSSLInfo struct {
	IsSSL      *int32  `json:"is_ssl,omitempty"`
	FullChain  *string `json:"full_chain,omitempty"`
	PrivateKey *string `json:"private_key,omitempty"`
}
