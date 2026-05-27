package cert

import (
	"crypto"
	"crypto/x509"

	"github.com/go-acme/lego/v5/certcrypto"
)

// 从 PEM 编码的证书字符串解析并返回一个 x509.Certificate 对象。
// PEM 内容可能是包含多张证书的证书链，但只返回第一个证书（即服务器证书）。
//
// 入参:
//   - certPEM: 证书 PEM 内容。
//
// 出参:
//   - cert: x509.Certificate 对象。
//   - err: 错误。
func ParseCertificateFromPEM(certPEM string) (_cert *x509.Certificate, _err error) {
	return certcrypto.ParsePEMCertificate([]byte(certPEM))
}

// 从 PEM 编码的私钥字符串解析并返回一个 crypto.PrivateKey 对象。
//
// 入参:
//   - privkeyPEM: 私钥 PEM 内容。
//
// 出参:
//   - privkey: crypto.PrivateKey 对象，可能是 rsa.PrivateKey、ecdsa.PrivateKey 或 ed25519.PrivateKey。
//   - err: 错误。
func ParsePrivateKeyFromPEM(privkeyPEM string) (_privkey crypto.PrivateKey, _err error) {
	return certcrypto.ParsePEMPrivateKey([]byte(privkeyPEM))
}
