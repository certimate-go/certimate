package cert

import (
	"bytes"
	"crypto/x509"
	"errors"
	"fmt"
	"time"

	"github.com/pavlo-v-chernykh/keystore-go/v4"
	"software.sslmate.com/src/go-pkcs12"
)

// 将 PEM 编码的证书字符串转换为 PFX 格式。
//
// 入参:
//   - certPEM: 证书 PEM 内容。
//   - privkeyPEM: 私钥 PEM 内容。
//   - pfxPassword: PFX 导出密码。
//
// 出参:
//   - data: PFX 格式的证书数据。
//   - err: 错误。
func TransformCertificateFromPEMToPFX(certPEM string, privkeyPEM string, pfxPassword string) ([]byte, error) {
	blocks := decodePEM([]byte(certPEM))

	certs := make([]*x509.Certificate, 0, len(blocks))
	for i, block := range blocks {
		if block.Type != "CERTIFICATE" {
			return nil, fmt.Errorf("invalid PEM block type at %d, expected 'CERTIFICATE', got '%s'", i, block.Type)
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}

	privkey, err := ParsePrivateKeyFromPEM(privkeyPEM)
	if err != nil {
		return nil, err
	}

	var pfxData []byte
	if len(certs) == 0 {
		return nil, errors.New("failed to decode certificate PEM")
	} else if len(certs) == 1 {
		pfxData, err = pkcs12.Legacy.Encode(privkey, certs[0], nil, pfxPassword)
	} else {
		pfxData, err = pkcs12.Legacy.Encode(privkey, certs[0], certs[1:], pfxPassword)
	}

	return pfxData, err
}

// 将 PEM 编码的证书字符串转换为 JKS 格式。
//
// 入参:
//   - certPEM: 证书 PEM 内容。
//   - privkeyPEM: 私钥 PEM 内容。
//   - jksAlias: JKS 别名。
//   - jksKeypass: JKS 密钥密码。
//   - jksStorepass: JKS 存储密码。
//
// 出参:
//   - data: JKS 格式的证书数据。
//   - err: 错误。
func TransformCertificateFromPEMToJKS(certPEM string, privkeyPEM string, jksAlias string, jksKeypass string, jksStorepass string) ([]byte, error) {
	certBlocks := decodePEM([]byte(certPEM))
	if len(certBlocks) == 0 {
		return nil, errors.New("failed to decode certificate PEM")
	}

	privkeyBlocks := decodePEM([]byte(privkeyPEM))
	if len(privkeyBlocks) == 0 {
		return nil, errors.New("failed to decode private key PEM")
	}

	entry := keystore.PrivateKeyEntry{
		CreationTime:     time.Now(),
		PrivateKey:       privkeyBlocks[0].Bytes,
		CertificateChain: make([]keystore.Certificate, len(certBlocks)),
	}
	for i, certBlock := range certBlocks {
		if certBlock.Type != "CERTIFICATE" {
			return nil, fmt.Errorf("invalid PEM block type at %d, expected 'CERTIFICATE', got '%s'", i, certBlock.Type)
		}

		entry.CertificateChain[i] = keystore.Certificate{
			Type:    "X509",
			Content: certBlock.Bytes,
		}
	}

	ks := keystore.New()
	if err := ks.SetPrivateKeyEntry(jksAlias, entry, []byte(jksKeypass)); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := ks.Store(&buf, []byte(jksStorepass)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
