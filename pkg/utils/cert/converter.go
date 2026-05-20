package cert

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// 将 x509.Certificate 对象转换为 PEM 编码的字符串。
//
// 入参:
//   - cert: x509.Certificate 对象。
//
// 出参:
//   - certPEM: 证书 PEM 内容。
//   - err: 错误。
func ConvertCertificateToPEM(cert *x509.Certificate) (_certPEM string, _err error) {
	if cert == nil {
		return "", fmt.Errorf("the input certificate is nil")
	}

	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}

	return string(pem.EncodeToMemory(block)), nil
}

// 将 ecdsa.PrivateKey 对象转换为 PEM 编码的字符串。
//
// 入参:
//   - privkey: ecdsa.PrivateKey 对象。
//
// 出参:
//   - privkeyPEM: 私钥 PEM 内容。
//   - err: 错误。
func ConvertECPrivateKeyToPEM(privkey *ecdsa.PrivateKey) (_privkeyPEM string, _err error) {
	if privkey == nil {
		return "", fmt.Errorf("the input private key is nil")
	}

	data, _err := x509.MarshalECPrivateKey(privkey)
	if _err != nil {
		return "", fmt.Errorf("failed to marshal EC private key: %w", _err)
	}

	block := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: data,
	}

	return string(pem.EncodeToMemory(block)), nil
}
