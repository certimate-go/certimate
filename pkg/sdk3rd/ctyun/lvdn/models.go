package lvdn

type Domain struct {
	Domain      string `json:"domain"`
	Cname       string `json:"cname"`
	ProductCode string `json:"product_code"`
	ProductName string `json:"product_name"`
	AreaScope   int32  `json:"area_scope"`
	Status      int32  `json:"status"`
}

type DomainDetail struct {
	Domain
	HttpsSwitch int32  `json:"https_switch"`
	CertName    string `json:"cert_name"`
}

type Cert struct {
	Id          int64    `json:"id"`
	Name        string   `json:"name"`
	CN          string   `json:"cn"`
	SANs        []string `json:"sans"`
	UsageMode   int32    `json:"usage_mode"`
	State       int32    `json:"state"`
	ExpiresTime int64    `json:"expires"`
	IssueTime   int64    `json:"issue"`
	Issuer      string   `json:"issuer"`
	CreatedTime int64    `json:"created"`
}

type CertDetail struct {
	Cert
	Certs string `json:"certs"`
	Key   string `json:"key"`
}
