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
	FILE_FORMAT_PEM = string(domain.CertificateFormatTypePEM)
	FILE_FORMAT_PFX = string(domain.CertificateFormatTypePFX)
	FILE_FORMAT_JKS = string(domain.CertificateFormatTypeJKS)
)
