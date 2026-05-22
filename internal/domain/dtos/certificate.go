package dtos

import (
	"github.com/certimate-go/certimate/internal/domain"
)

type CertificateDownloadReq struct {
	CertificateId string                       `json:"-"`
	FileFormat    domain.CertificateFormatType `json:"format"`
}

type CertificateDownloadResp struct {
	ZipBytes []byte `json:"zipBytes"`
}

type CertificateRevokeReq struct {
	CertificateId string `json:"-"`
}

type CertificateRevokeResp struct{}
