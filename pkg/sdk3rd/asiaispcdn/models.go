package asiaispcdn

type Domain struct {
	Domain          string `json:"domain"`
	ServiceType     string `json:"serviceType"`
	Scope           string `json:"scope"`
	Cname           string `json:"cname"`
	ICP             string `json:"icp"`
	Protocol        string `json:"protocol"`
	CertId          int64  `json:"certId"`
	OriginProtocol  string `json:"originProtocol"`
	OriginHost      string `json:"originHost"`
	OriginType      string `json:"originType"`
	OriginAddr      string `json:"originAddr"`
	OriginBackAddr  string `json:"originBackAddr"`
	DomainStatus    int32  `json:"domainStatus"`
	OperatingStatus int32  `json:"operatingStatus"`
}

type Certificate struct {
	CertId string `json:"certId"`
	Name   string `json:"name"`
}

type CertificateDetail struct {
	Certificate

	PrivateKey  string `json:"privateKey"`
	PublicKey   string `json:"publicKey"`
	SubjectName string `json:"subjectName"`
	Brand       string `json:"brand"`
	IssueTime   string `json:"issueTime"`
	ExpireTime  string `json:"expireTime"`
}
