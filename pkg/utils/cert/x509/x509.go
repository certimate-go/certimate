package x509

import (
	"crypto/x509"
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
		for _, dnsName := range cert.DNSNames {
			sans = append(sans, dnsName)
		}
		for _, ipAddr := range cert.IPAddresses {
			sans = append(sans, ipAddr.String())
		}
		for _, email := range cert.EmailAddresses {
			sans = append(sans, email)
		}
		for _, uri := range cert.URIs {
			if uri != nil {
				sans = append(sans, uri.String())
			}
		}
	}

	return sans
}
