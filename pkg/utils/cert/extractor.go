package cert

import (
	"encoding/pem"
	"errors"
	"fmt"
)

// 从 PEM 编码的证书字符串解析并提取服务器证书和中间证书。
//
// 入参:
//   - certPEM: 证书 PEM 内容。
//
// 出参:
//   - serverCertPEM: 服务器证书的 PEM 内容。
//   - intermediaCertPEM: 中间证书的 PEM 内容。
//   - err: 错误。
func ExtractCertificatesFromPEM(certPEM string) (_serverCertPEM string, _intermediaCertPEM string, _err error) {
	blocks := decodePEM([]byte(certPEM))
	for i, block := range blocks {
		if block.Type != "CERTIFICATE" {
			return "", "", fmt.Errorf("invalid PEM block type at %d, expected 'CERTIFICATE', got '%s'", i, block.Type)
		}
	}

	serverCertPEM := ""
	intermediaCertPEM := ""

	if len(blocks) == 0 {
		return "", "", errors.New("failed to decode PEM block")
	}

	if len(blocks) > 0 {
		serverCertPEM = string(pem.EncodeToMemory(blocks[0]))
	}

	if len(blocks) > 1 {
		for i := 1; i < len(blocks); i++ {
			intermediaCertPEM += string(pem.EncodeToMemory(blocks[i]))
		}
	}

	return serverCertPEM, intermediaCertPEM, nil
}
