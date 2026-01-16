package s3

import (
	"github.com/certimate-go/certimate/internal/domain"
)

const (
	OUTPUT_FORMAT_PEM = string(domain.CertificateFormatTypePEM)
	OUTPUT_FORMAT_PFX = string(domain.CertificateFormatTypePFX)
	OUTPUT_FORMAT_JKS = string(domain.CertificateFormatTypeJKS)
)
