package onepanel

type Website struct {
	ID            int64  `json:"id"`
	Alias         string `json:"alias"`
	PrimaryDomain string `json:"primaryDomain"`
	Protocol      string `json:"protocol"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	SitePath      string `json:"sitePath"`
	Remark        string `json:"remark"`
	SSLStatus     string `json:"sslStatus,omitempty"`
	SSLExpireDate string `json:"sslExpireDate,omitempty"`
	WebsiteSSLID  int64  `json:"webSiteSSLId,omitempty"`
	UpdatedAt     string `json:"updatedAt"`
	CreatedAt     string `json:"createdAt"`
}

type WebsiteDetail struct {
	Website
	Domains []*WebsiteDomainConfig `json:"domains"`
}

type WebsiteDomainConfig struct {
	ID        int64  `json:"id"`
	Domain    string `json:"domain"`
	Port      int32  `json:"port"`
	SSL       bool   `json:"ssl"`
	UpdatedAt string `json:"updatedAt"`
	CreatedAt string `json:"createdAt"`
}

type WebsiteHTTPSConfig struct {
	Enable       bool     `json:"enable"`
	WebsiteSSLID int64    `json:"websiteSSLId"`
	HttpConfig   string   `json:"httpConfig"`
	SSLProtocol  []string `json:"SSLProtocol"`
	Algorithm    string   `json:"algorithm"`
	Hsts         bool     `json:"hsts"`
}

type SSLCertificate struct {
	ID          int64  `json:"id"`
	PEM         string `json:"pem"`
	PrivateKey  string `json:"privateKey"`
	Domains     string `json:"domains"`
	Description string `json:"description"`
	Status      string `json:"status"`
	UpdatedAt   string `json:"updatedAt"`
	CreatedAt   string `json:"createdAt"`
}
