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
				var raw asn1.RawValue
				_, err := asn1.Unmarshal(ext.Value, &raw)
				if err != nil {
					continue
				}

				var seq asn1.RawValue
				if _, err := asn1.Unmarshal(raw.Bytes, &seq); err != nil {
					continue
				}

				switch seq.Tag {
				case sanGeneralNameTagIP:
					// IPv4 地址需要单独处理，否则直接转换为字符串会得到乱码
					var ip net.IP = seq.Bytes
					sans = append(sans, ip.String())

				case sanGeneralNameTagEmail, sanGeneralNameTagDNS, sanGeneralNameTagURI:
					sans = append(sans, string(seq.Bytes))

				default:
					// 忽略其他非 Critical 的 GeneralName​
				}
			}
		}
	}

	return sans
}
