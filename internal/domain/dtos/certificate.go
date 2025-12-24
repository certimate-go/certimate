package dtos

type CertificateArchiveFileReq struct {
	CertificateId     string `json:"-"`
	CertificateFormat string `json:"format"`
}

type CertificateArchiveFileResp struct {
	FileBytes  []byte `json:"fileBytes"`
	FileFormat string `json:"fileFormat"`
}

type CertificateRevokeReq struct {
	CertificateId string `json:"-"`
}

type CertificateRevokeResp struct{}
