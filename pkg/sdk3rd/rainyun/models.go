package rainyun

type SSLCertificate struct {
	ID         int64  `json:"ID"`
	UID        int64  `json:"UID"`
	Domain     string `json:"Domain"`
	Issuer     string `json:"Issuer"`
	StartDate  int64  `json:"StartDate"`
	ExpireDate int64  `json:"ExpDate"`
	UploadTime int64  `json:"UploadTime"`
}

type SSLCertificateDetail struct {
	Cert       string `json:"Cert"`
	Key        string `json:"Key"`
	Domain     string `json:"DomainName"`
	Issuer     string `json:"Issuer"`
	StartDate  int64  `json:"StartDate"`
	ExpireDate int64  `json:"ExpDate"`
	RemainDays int64  `json:"RemainDays"`
}
