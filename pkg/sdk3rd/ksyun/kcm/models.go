package kcm

type UserCertificate struct {
	CertId            string   `json:"CertID"`
	CertName          string   `json:"CertName"`
	MainDomain        string   `json:"MainDomain"`
	Domains           []string `json:"Domains"`
	AdditionalDomains []string `json:"AdditionalDomains"`
	Brand             string   `json:"Brand"`
	CA                string   `json:"CA"`
	Level             string   `json:"Level"`
	FingerPrint       string   `json:"FingerPrint"`
	DomainCount       int32    `json:"DomainCount"`
	WildcardCount     int32    `json:"WildcardCount"`
	IssueTime         string   `json:"IssueTime"`
	ExpireTime        string   `json:"ExpireTime"`
	Source            string   `json:"Source"`
}

type LBCertificate struct {
	CertificateId    string `json:"CertificateId"`
	CertificateName  string `json:"CertificateName"`
	CommonName       string `json:"CommonName"`
	CertAuthority    string `json:"CertAuthority"`
	CertType         string `json:"CertType"`
	CertificateType  string `json:"CertificateType"`
	PublicKey        string `json:"PublicKey"`
	ExpireTime       string `json:"ExpireTime"`
	CreateTime       string `json:"CreateTime"`
	Source           string `json:"Source"`
	SSLCertificateId string `json:"SslCertificateId,omitempty"`
}
