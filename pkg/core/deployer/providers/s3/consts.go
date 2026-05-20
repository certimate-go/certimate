package s3

import (
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/local/shared"
)

const (
	FILE_FORMAT_PEM = shared.FILE_FORMAT_PEM
	FILE_FORMAT_PFX = shared.FILE_FORMAT_PFX
	FILE_FORMAT_JKS = shared.FILE_FORMAT_JKS
)

const (
	PFX_ENCODER_LEGACYRC2  = shared.PFX_ENCODER_LEGACYRC2
	PFX_ENCODER_LEGACYDES  = shared.PFX_ENCODER_LEGACYDES
	PFX_ENCODER_MODERN2023 = shared.PFX_ENCODER_MODERN2023
	PFX_ENCODER_MODERN2026 = shared.PFX_ENCODER_MODERN2026
)
