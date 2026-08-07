package cms

type Certificate struct {
	Id                  string `json:"id"`
	Origin              string `json:"origin"`
	Type                string `json:"type"`
	ResourceId          string `json:"resourceId"`
	ResourceType        string `json:"resourceType"`
	CertificateId       string `json:"certificateId"`
	CertificateMode     string `json:"certificateMode"`
	Name                string `json:"name"`
	Status              string `json:"status"`
	DetailStatus        string `json:"detailStatus"`
	ManagedStatus       string `json:"managedStatus"`
	Fingerprint         string `json:"fingerprint"`
	IssueTime           string `json:"issueTime"`
	ExpireTime          string `json:"expireTime"`
	DomainType          string `json:"domainType"`
	DomainName          string `json:"domainName"`
	EncryptionStandard  string `json:"encryptionStandard"`
	EncryptionAlgorithm string `json:"encryptionAlgorithm"`
	CreateTime          string `json:"createTime"`
	UpdateTime          string `json:"updateTime"`
}
