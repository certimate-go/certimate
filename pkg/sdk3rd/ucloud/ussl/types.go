package ussl

type CertificateListItem struct {
	CertificateID     int
	CertificateSN     string
	CertificateCat    string
	Mode              string
	Domains           string
	Brand             string
	ValidityPeriod    int
	Type              string
	NotBefore         int64
	NotAfter          int64
	AlarmState        int
	State             string
	StateCode         string
	Name              string
	MaxDomainsCount   int
	DomainsCount      int
	CaChannel         string
	CSRAlgorithms     []CSRAlgorithmInfo
	TopOrganizationID int
	OrganizationID    int
	IsFree            int
	YearOfValidity    int
	Channel           int
	CreateTime        int
	CertificateUrl    string
}

type CSRAlgorithmInfo struct {
	Algorithm       string
	AlgorithmOption []string
}

type CertificateInfo struct {
	Type            string
	CertificateID   int
	CertificateType string
	CaOrganization  string
	Algorithm       string
	ValidityPeriod  int
	State           string
	StateCode       string
	Name            string
	Brand           string
	Domains         string
	DomainsCount    int
	Mode            string
	CSROnline       int
	CSR             string
	CSRKeyParameter string
	CSREncryptAlgo  string
	IssuedDate      int
	ExpiredDate     int
}
