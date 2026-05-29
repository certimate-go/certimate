package cert

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
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

// 将 rsa.PrivateKey 或 ecdsa.PrivateKey 对象转换为 PEM 编码的字符串。
//
// 入参:
//   - privkey: rsa.PrivateKey 或 ecdsa.PrivateKey 对象。
//   - pkcs8: 是否使用 PKCS#8 格式编码。
//
// 出参:
//   - privkeyPEM: 私钥 PEM 内容。
//   - err: 错误。
func ConvertPrivateKeyToPEM(privkey crypto.PrivateKey, pkcs8 bool) (_privkeyPEM string, _err error) {
	if privkey == nil {
		return "", fmt.Errorf("the input private key is nil")
	}

	switch t := privkey.(type) {
	case *rsa.PrivateKey:
		return ConvertRSAPrivateKeyToPEM(t, pkcs8)

	case *ecdsa.PrivateKey:
		return ConvertECPrivateKeyToPEM(t, pkcs8)
	}

	return "", fmt.Errorf("unknown private key type")
}

// 将 rsa.PrivateKey 对象转换为 PEM 编码的字符串。
//
// 入参:
//   - privkey: rsa.PrivateKey 对象。
//   - pkcs8: 是否使用 PKCS#8 格式编码。否则，使用 PKCS#1 格式编码。
//
// 出参:
//   - privkeyPEM: 私钥 PEM 内容。
//   - err: 错误。
func ConvertRSAPrivateKeyToPEM(privkey *rsa.PrivateKey, pkcs8 bool) (_privkeyPEM string, _err error) {
	if privkey == nil {
		return "", fmt.Errorf("the input private key is nil")
	}

	var data []byte
	if pkcs8 {
		data, _err = x509.MarshalPKCS8PrivateKey(privkey)
		if _err != nil {
			return "", fmt.Errorf("failed to marshal RSA private key: %w", _err)
		}
	} else {
		data = x509.MarshalPKCS1PrivateKey(privkey)
		if data == nil {
			_err = fmt.Errorf("failed to marshal RSA private key")
			return "", _err
		}
	}

	var block *pem.Block
	if pkcs8 {
		block = &pem.Block{Type: "PRIVATE KEY", Bytes: data}
	} else {
		block = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: data}
	}

	return string(pem.EncodeToMemory(block)), nil
}

// 将 ecdsa.PrivateKey 对象转换为 PEM 编码的字符串。
//
// 入参:
//   - privkey: ecdsa.PrivateKey 对象。
//   - pkcs8: 是否使用 PKCS#8 格式编码。否则，使用 SEC1 格式编码。
//
// 出参:
//   - privkeyPEM: 私钥 PEM 内容。
//   - err: 错误。
func ConvertECPrivateKeyToPEM(privkey *ecdsa.PrivateKey, pkcs8 bool) (_privkeyPEM string, _err error) {
	if privkey == nil {
		return "", fmt.Errorf("the input private key is nil")
	}

	var data []byte
	if pkcs8 {
		data, _err = x509.MarshalPKCS8PrivateKey(privkey)
		if _err != nil {
			return "", fmt.Errorf("failed to marshal EC private key: %w", _err)
		}
	} else {
		data, _err = x509.MarshalECPrivateKey(privkey)
		if _err != nil {
			return "", fmt.Errorf("failed to marshal EC private key: %w", _err)
		}
	}

	var block *pem.Block
	if pkcs8 {
		block = &pem.Block{Type: "PRIVATE KEY", Bytes: data}
	} else {
		block = &pem.Block{Type: "EC PRIVATE KEY", Bytes: data}
	}

	return string(pem.EncodeToMemory(block)), nil
}
