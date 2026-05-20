package local

import (
	"fmt"
	"strings"

	"software.sslmate.com/src/go-pkcs12"
)

func ResolvePfxEncoder(encoderName string) (*pkcs12.Encoder, error) {
	var encoder *pkcs12.Encoder

	if encoderName != "" {
		if strings.EqualFold(encoderName, PFX_ENCODER_LEGACYRC2) {
			encoder = pkcs12.LegacyRC2
		} else if strings.EqualFold(encoderName, PFX_ENCODER_LEGACYDES) {
			encoder = pkcs12.LegacyDES
		} else if strings.EqualFold(encoderName, PFX_ENCODER_MODERN2023) {
			encoder = pkcs12.Modern2023
		} else if strings.EqualFold(encoderName, PFX_ENCODER_MODERN2026) {
			encoder = pkcs12.Modern2026
		} else {
			return nil, fmt.Errorf("unsupported encoder name: '%s'", encoderName)
		}
	}

	return encoder, nil
}
