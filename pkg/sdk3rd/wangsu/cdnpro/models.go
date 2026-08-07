package cdnpro

type CertificateVersion struct {
	Comments           *string                               `json:"comments,omitempty"`
	PrivateKey         *string                               `json:"privateKey,omitempty"`
	Certificate        *string                               `json:"certificate,omitempty"`
	ChainCert          *string                               `json:"chainCert,omitempty"`
	IdentificationInfo *CertificateVersionIdentificationInfo `json:"identificationInfo,omitempty"`
}

type CertificateVersionIdentificationInfo struct {
	Country                 *string   `json:"country,omitempty"`
	State                   *string   `json:"state,omitempty"`
	City                    *string   `json:"city,omitempty"`
	Company                 *string   `json:"company,omitempty"`
	Department              *string   `json:"department,omitempty"`
	CommonName              *string   `json:"commonName,omitempty"`
	Email                   *string   `json:"email,omitempty"`
	SubjectAlternativeNames []*string `json:"subjectAlternativeNames,omitempty"`
}

type HostnameProperty struct {
	PropertyId    string  `json:"propertyId"`
	Version       int32   `json:"version"`
	CertificateId *string `json:"certificateId,omitempty"`
}

type DeploymentTaskAction struct {
	Action        *string `json:"action,omitempty"`
	PropertyId    *string `json:"propertyId,omitempty"`
	CertificateId *string `json:"certificateId,omitempty"`
	Version       *int32  `json:"version,omitempty"`
}
