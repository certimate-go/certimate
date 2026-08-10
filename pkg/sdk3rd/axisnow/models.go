package axisnow

type Certificate struct {
	UUID            string             `json:"uuid,omitempty"`
	Name            string             `json:"name,omitempty"`
	Type            string             `json:"type,omitempty"`
	SubjectName     []string           `json:"subject_name,omitempty"`
	Issuer          CertificateIssuer  `json:"issuer,omitempty"`
	Subject         CertificateSubject `json:"subject,omitempty"`
	IssueTime       int64              `json:"issue_time,omitempty"`
	ExpireTime      int64              `json:"expire_time,omitempty"`
	Certificate     string             `json:"certificate,omitempty"`
	ReferencedCount int                `json:"referenced_count,omitempty"`
	CreatedAt       string             `json:"created_at,omitempty"`
	UpdatedAt       string             `json:"updated_at,omitempty"`
}

type CertificateIssuer struct {
	Country          string `json:"country,omitempty"`
	Organization     string `json:"organization,omitempty"`
	OrganizationUnit string `json:"organization_unit,omitempty"`
	CommonName       string `json:"common_name,omitempty"`
}

type CertificateSubject struct {
	Organization     string `json:"organization,omitempty"`
	OrganizationUnit string `json:"organization_unit,omitempty"`
	CommonName       string `json:"common_name,omitempty"`
}
