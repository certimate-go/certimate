package ao

type Domain struct {
	Domain      string `json:"domain"`
	Cname       string `json:"cname"`
	ProductCode string `json:"product_code"`
	ProductName string `json:"product_name"`
	Status      int32  `json:"status"`
	AreaScope   int32  `json:"area_scope"`
}

type DomainOriginConfig struct {
	Origin string `json:"origin"`
	Role   string `json:"role"`
	Weight string `json:"weight"`
}

type DomainOriginConfigWithWeight struct {
	Origin string `json:"origin"`
	Role   string `json:"role"`
	Weight int32  `json:"weight"`
}

type DomainHttpsBasicConfig struct {
	HttpsForce  string `json:"https_force,omitempty"`
	ForceStatus string `json:"force_status,omitempty"`
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
