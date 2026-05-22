package s3

import (
	"github.com/certimate-go/certimate/internal/domain"
	xcertpfx "github.com/certimate-go/certimate/pkg/utils/cert/pfx"
)

const (
	FILE_FORMAT_PEM = string(domain.CertificateFormatTypePEM)
	FILE_FORMAT_PFX = string(domain.CertificateFormatTypePFX)
	FILE_FORMAT_JKS = string(domain.CertificateFormatTypeJKS)
)

const (
	PFX_ENCODER_LEGACYRC2  = string(xcertpfx.EncoderNameLegacyRC2)
	PFX_ENCODER_LEGACYDES  = string(xcertpfx.EncoderNameLegacyDES)
	PFX_ENCODER_MODERN2023 = string(xcertpfx.EncoderNameModern2023)
	PFX_ENCODER_MODERN2026 = string(xcertpfx.EncoderNameModern2026)
)
