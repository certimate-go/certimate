package domain

import (
	"crypto/x509"
	"fmt"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"

	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xcertkey "github.com/certimate-go/certimate/pkg/utils/cert/key"
	xcertx509 "github.com/certimate-go/certimate/pkg/utils/cert/x509"
)

const CollectionNameCertificate = "certificate"

type Certificate struct {
	Meta
	Source            CertificateSourceType       `db:"source"            json:"source"`
	SubjectAltNames   string                      `db:"subjectAltNames"   json:"subjectAltNames"`
	SerialNumber      string                      `db:"serialNumber"      json:"serialNumber"`
	Certificate       string                      `db:"certificate"       json:"certificate"`
	PrivateKey        string                      `db:"privateKey"        json:"privateKey"`
	IssuerOrg         string                      `db:"issuerOrg"         json:"issuerOrg"`
	IssuerCertificate string                      `db:"issuerCertificate" json:"issuerCertificate"`
	KeyAlgorithm      CertificateKeyAlgorithmType `db:"keyAlgorithm"      json:"keyAlgorithm"`
	ValidityNotBefore time.Time                   `db:"validityNotBefore" json:"validityNotBefore"`
	ValidityNotAfter  time.Time                   `db:"validityNotAfter"  json:"validityNotAfter"`
	ValidityInterval  int32                       `db:"validityInterval"  json:"validityInterval"`
	ACMEAcctUrl       string                      `db:"acmeAcctUrl"       json:"acmeAcctUrl"`
	ACMECertUrl       string                      `db:"acmeCertUrl"       json:"acmeCertUrl"`
	IsRenewed         bool                        `db:"isRenewed"         json:"isRenewed"`
	IsRevoked         bool                        `db:"isRevoked"         json:"isRevoked"`
	WorkflowId        string                      `db:"workflowRef"       json:"workflowId"`
	WorkflowRunId     string                      `db:"workflowRunRef"    json:"workflowRunId"`
	WorkflowNodeId    string                      `db:"workflowNodeId"    json:"workflowNodeId"`
	DeletedAt         *time.Time                  `db:"deleted" json:"deleted"`
}

func (c *Certificate) PopulateFromX509(certX509 *x509.Certificate) *Certificate {
	c.SubjectAltNames = strings.Join(xcertx509.GetSubjectAltNames(certX509), ";")
	c.SerialNumber = strings.ToUpper(certX509.SerialNumber.Text(16))
	c.IssuerOrg = strings.Join(certX509.Issuer.Organization, ";")
	c.ValidityNotBefore = certX509.NotBefore
	c.ValidityNotAfter = certX509.NotAfter
	c.ValidityInterval = int32(certX509.NotAfter.Sub(certX509.NotBefore).Seconds())

	keyAlgorithm, keySize, _ := xcertkey.GetPublicKeyAlgorithm(certX509.PublicKey)
	switch keyAlgorithm {
	case x509.RSA:
		c.KeyAlgorithm = CertificateKeyAlgorithmType(fmt.Sprintf("RSA%d", keySize))
	case x509.ECDSA:
		c.KeyAlgorithm = CertificateKeyAlgorithmType(fmt.Sprintf("EC%d", keySize))
	case x509.Ed25519:
		c.KeyAlgorithm = CertificateKeyAlgorithmType("Ed25519")
	default:
		c.KeyAlgorithm = CertificateKeyAlgorithmType("")
	}

	return c
}

func (c *Certificate) PopulateFromPEM(certPEM, privkeyPEM string) *Certificate {
	c.Certificate = certPEM
	c.PrivateKey = privkeyPEM

	_, issuerCertPEM, _ := xcert.ExtractCertificatesFromPEM(certPEM)
	c.IssuerCertificate = issuerCertPEM

	certX509, _ := xcert.ParseCertificateFromPEM(certPEM)
	if certX509 != nil {
		return c.PopulateFromX509(certX509)
	}

	return c
}

type CertificateSourceType string

const (
	CertificateSourceTypeRequest = CertificateSourceType("request")
	CertificateSourceTypeUpload  = CertificateSourceType("upload")
)

type CertificateKeyAlgorithmType string

const (
	CertificateKeyAlgorithmTypeRSA2048 = CertificateKeyAlgorithmType("RSA2048")
	CertificateKeyAlgorithmTypeRSA3072 = CertificateKeyAlgorithmType("RSA3072")
	CertificateKeyAlgorithmTypeRSA4096 = CertificateKeyAlgorithmType("RSA4096")
	CertificateKeyAlgorithmTypeRSA8192 = CertificateKeyAlgorithmType("RSA8192")
	CertificateKeyAlgorithmTypeEC256   = CertificateKeyAlgorithmType("EC256")
	CertificateKeyAlgorithmTypeEC384   = CertificateKeyAlgorithmType("EC384")
	CertificateKeyAlgorithmTypeEC512   = CertificateKeyAlgorithmType("EC512")
)

func (t CertificateKeyAlgorithmType) KeyType() (certcrypto.KeyType, error) {
	keyTypeMap := map[CertificateKeyAlgorithmType]certcrypto.KeyType{
		CertificateKeyAlgorithmTypeRSA2048: certcrypto.RSA2048,
		CertificateKeyAlgorithmTypeRSA3072: certcrypto.RSA3072,
		CertificateKeyAlgorithmTypeRSA4096: certcrypto.RSA4096,
		CertificateKeyAlgorithmTypeRSA8192: certcrypto.RSA8192,
		CertificateKeyAlgorithmTypeEC256:   certcrypto.EC256,
		CertificateKeyAlgorithmTypeEC384:   certcrypto.EC384,
	}

	if keyType, ok := keyTypeMap[t]; ok {
		return keyType, nil
	}

	return certcrypto.RSA2048, fmt.Errorf("unsupported key algorithm type: '%s'", t)
}

type CertificateFormatType string

const (
	CertificateFormatTypePEM CertificateFormatType = "PEM"
	CertificateFormatTypePFX CertificateFormatType = "PFX"
	CertificateFormatTypeJKS CertificateFormatType = "JKS"
)
