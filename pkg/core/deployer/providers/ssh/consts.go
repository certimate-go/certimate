package ssh

import (
	"github.com/certimate-go/certimate/internal/domain"
)

const (
	FILE_FORMAT_PEM = string(domain.CertificateFormatTypePEM)
	FILE_FORMAT_PFX = string(domain.CertificateFormatTypePFX)
	FILE_FORMAT_JKS = string(domain.CertificateFormatTypeJKS)
)
