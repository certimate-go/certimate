package local

import (
	"github.com/certimate-go/certimate/internal/domain"
)

const (
	SHELL_ENV_SH         = "sh"
	SHELL_ENV_CMD        = "cmd"
	SHELL_ENV_POWERSHELL = "powershell"
)

const (
	OUTPUT_FORMAT_PEM = string(domain.CertificateFormatTypePEM)
	OUTPUT_FORMAT_PFX = string(domain.CertificateFormatTypePFX)
	OUTPUT_FORMAT_JKS = string(domain.CertificateFormatTypeJKS)
)
