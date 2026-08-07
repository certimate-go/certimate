package cert

import (
	"encoding/pem"
	"fmt"
)

// 从 PEM 编码的证书字符串解析并提取叶子证书和中间证书。
//
// 入参:
//   - certPEM: 证书 PEM 内容。
//
// 出参:
//   - leafCertPEM: 叶子证书的 PEM 内容。
//   - intermediateCertPEM: 中间证书的 PEM 内容。
//   - err: 错误。
func ExtractCertificatesFromPEM(certPEM string) (_leafCertPEM string, _intermediateCertPEM string, _err error) {
	blocks := decodePEMBlocks([]byte(certPEM))
	for i, block := range blocks {
		if block.Type != "CERTIFICATE" {
			return "", "", fmt.Errorf("invalid PEM block type at %d, expected 'CERTIFICATE', got '%s'", i, block.Type)
		}
	}

	_leafCertPEM = ""
	_intermediateCertPEM = ""

	if len(blocks) == 0 {
		return "", "", fmt.Errorf("failed to decode PEM block")
	}

	if len(blocks) > 0 {
		_leafCertPEM = string(pem.EncodeToMemory(blocks[0]))
	}

	if len(blocks) > 1 {
		for i := 1; i < len(blocks); i++ {
			_intermediateCertPEM += string(pem.EncodeToMemory(blocks[i]))
		}
	}

	return _leafCertPEM, _intermediateCertPEM, nil
}
