package pfx

import (
	"fmt"
	"strings"

	"software.sslmate.com/src/go-pkcs12"
)

const (
	EncoderNameLegacyRC2  = "LegacyRC2"
	EncoderNameLegacyDES  = "LegacyDES"
	EncoderNameModern2023 = "Modern2023"
	EncoderNameModern2026 = "Modern2026"
)

func ResolvePfxEncoder(encoderName string) (*pkcs12.Encoder, error) {
	var encoder *pkcs12.Encoder

	if encoderName != "" {
		if strings.EqualFold(encoderName, EncoderNameLegacyRC2) {
			encoder = pkcs12.LegacyRC2
		} else if strings.EqualFold(encoderName, EncoderNameLegacyDES) {
			encoder = pkcs12.LegacyDES
		} else if strings.EqualFold(encoderName, EncoderNameModern2023) {
			encoder = pkcs12.Modern2023
		} else if strings.EqualFold(encoderName, EncoderNameModern2026) {
			encoder = pkcs12.Modern2026
		} else {
			return nil, fmt.Errorf("unknown encoder name: %s", encoderName)
		}
	}

	return encoder, nil
}
