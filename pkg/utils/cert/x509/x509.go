package x509

import (
	"crypto/x509"
	"encoding/asn1"
	"net"
)

var oidSubjectAlternativeNameExtension = asn1.ObjectIdentifier{2, 5, 29, 17}

const (
	sanGeneralNameTagEmail = 1
	sanGeneralNameTagDNS   = 2
	sanGeneralNameTagURI   = 6
	sanGeneralNameTagIP    = 7
)

// 返回指定 x509.Certificate 对象的主题名称。
// 如果主题名称为空，则返回第一个主题替代名称。
//
// 入参：
//   - cert: x509.Certificate 对象。
//
// 出参：
//   - 主题名称。
func GetSubjectCommonName(cert *x509.Certificate) string {
	if cert != nil {
		if cert.Subject.CommonName != "" {
			return cert.Subject.CommonName
		}

		sans := GetSubjectAltNames(cert)
		if len(sans) > 0 {
			return sans[0]
		}
	}

	return ""
}

// 返回指定 x509.Certificate 对象的主题替代名称。
//
// 入参：
//   - cert: x509.Certificate 对象。
//
// 出参：
//   - 主题替代名称的字符串切片。
func GetSubjectAltNames(cert *x509.Certificate) []string {
	sans := make([]string, 0)

	if cert != nil {
		// 注意，这里不直接使用 `DNSNames`、`IPAddresses` 等字段，以保证原始顺序不变
		for _, ext := range cert.Extensions {
			if ext.Id.Equal(oidSubjectAlternativeNameExtension) {
				var seq asn1.RawValue
				if _, err := asn1.Unmarshal(ext.Value, &seq); err != nil {
					continue
				}

				rest := seq.Bytes
				for len(rest) > 0 {
					var name asn1.RawValue
					var err error
					rest, err = asn1.Unmarshal(rest, &name)
					if err != nil {
						break
					}

					switch name.Tag {
					case sanGeneralNameTagIP:
						var ip net.IP = name.Bytes
						sans = append(sans, ip.String())

					case sanGeneralNameTagEmail, sanGeneralNameTagDNS, sanGeneralNameTagURI:
						sans = append(sans, string(name.Bytes))

					default:
						// 忽略其他非 Critical 的 GeneralName​
					}
				}
			}
		}
	}

	return sans
}
