package cert

import (
	"crypto/x509"
	"encoding/pem"
)

// 比较两个 x509.Certificate 对象，判断它们是否是同一张证书。
// 注意，这不是精确比较，而只是基于证书序列号和数字签名的快速判断，但对于权威 CA 签发的证书来说不会存在误判。
//
// 入参:
//   - a: 待比较的第一个 x509.Certificate 对象。
//   - b: 待比较的第二个 x509.Certificate 对象。
//
// 出参:
//   - 是否相同。
func EqualCertificate(a, b *x509.Certificate) bool {
	if a == nil || b == nil {
		return false
	}

	return string(a.Signature) == string(b.Signature) &&
		a.SignatureAlgorithm == b.SignatureAlgorithm &&
		a.SerialNumber.String() == b.SerialNumber.String() &&
		a.Issuer.SerialNumber == b.Issuer.SerialNumber &&
		a.Subject.SerialNumber == b.Subject.SerialNumber
}

func decodePEM(data []byte) []*pem.Block {
	blocks := make([]*pem.Block, 0)
	for {
		block, rest := pem.Decode(data)
		if block == nil {
			break
		}

		blocks = append(blocks, block)
		data = rest
	}

	return blocks
}
