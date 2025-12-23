package dtos

type CertificateArchiveFileReq struct {
	CertificateId string `json:"-"`
	Format        string `json:"format"`
}

type CertificateArchiveFileResp struct {
	FileBytes  []byte `json:"fileBytes"`
	FileFormat string `json:"fileFormat"`
}

type CertificateRevokeReq struct {
	CertificateId string `json:"-"`
}

type CertificateRevokeResp struct{}

type CertificateValidateCertificateReq struct {
	Certificate string `json:"certificate"`
}

type CertificateValidateCertificateResp struct {
	IsValid         bool   `json:"isValid"`
	SubjectAltNames string `json:"subjectAltNames,omitempty"`
}

type CertificateValidatePrivateKeyReq struct {
	PrivateKey string `json:"privateKey"`
}

type CertificateValidatePrivateKeyResp struct {
	IsValid      bool   `json:"isValid"`
	KeyAlgorithm string `json:"keyAlgorithm,omitempty"`
}
