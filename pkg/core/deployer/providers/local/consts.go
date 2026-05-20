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

const (
	PFX_ENCODER_LEGACYRC2  = "LegacyRC2"
	PFX_ENCODER_LEGACYDES  = "LegacyDES"
	PFX_ENCODER_MODERN2023 = "Modern2023"
	PFX_ENCODER_MODERN2026 = "Modern2026"
)
