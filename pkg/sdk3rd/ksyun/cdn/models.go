package cdn

type CDNDomain struct {
	Region       string `json:"Region"`
	DomainId     string `json:"DomainId"`
	DomainName   string `json:"DomainName"`
	Description  string `json:"Description"`
	Cname        string `json:"Cname"`
	CdnType      string `json:"CdnType"`
	CdnSubType   string `json:"CdnSubType"`
	DomainStatus string `json:"DomainStatus"`
	CreatedTime  string `json:"CreatedTime"`
	ModifiedTime string `json:"ModifiedTime"`
}
