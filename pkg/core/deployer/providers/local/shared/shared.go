package shared

import (
	"software.sslmate.com/src/go-pkcs12"

	"github.com/certimate-go/certimate/pkg/core/deployer/providers/local"
)

const (
	FILE_FORMAT_PEM = local.FILE_FORMAT_PEM
	FILE_FORMAT_PFX = local.FILE_FORMAT_PFX
	FILE_FORMAT_JKS = local.FILE_FORMAT_JKS
)

const (
	PFX_ENCODER_LEGACYRC2  = local.PFX_ENCODER_LEGACYRC2
	PFX_ENCODER_LEGACYDES  = local.PFX_ENCODER_LEGACYDES
	PFX_ENCODER_MODERN2023 = local.PFX_ENCODER_MODERN2023
	PFX_ENCODER_MODERN2026 = local.PFX_ENCODER_MODERN2026
)

func ResolvePfxEncoder(encoderName string) (*pkcs12.Encoder, error) {
	return local.ResolvePfxEncoder(encoderName)
}
