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
	Source            CertificateSourceType       `json:"source" db:"source"`
	SubjectAltNames   string                      `json:"subjectAltNames" db:"subjectAltNames"`
	SerialNumber      string                      `json:"serialNumber" db:"serialNumber"`
	Certificate       string                      `json:"certificate" db:"certificate"`
	PrivateKey        string                      `json:"privateKey" db:"privateKey"`
	IssuerOrg         string                      `json:"issuerOrg" db:"issuerOrg"`
	IssuerCertificate string                      `json:"issuerCertificate" db:"issuerCertificate"`
	KeyAlgorithm      CertificateKeyAlgorithmType `json:"keyAlgorithm" db:"keyAlgorithm"`
	ValidityNotBefore time.Time                   `json:"validityNotBefore" db:"validityNotBefore"`
	ValidityNotAfter  time.Time                   `json:"validityNotAfter" db:"validityNotAfter"`
	ValidityInterval  int32                       `json:"validityInterval" db:"validityInterval"`
	ACMEAcctUrl       string                      `json:"acmeAcctUrl" db:"acmeAcctUrl"`
	ACMECertUrl       string                      `json:"acmeCertUrl" db:"acmeCertUrl"`
	ACMECertStableUrl string                      `json:"acmeCertStableUrl" db:"acmeCertStableUrl"`
	IsRenewed         bool                        `json:"isRenewed" db:"isRenewed"`
	IsRevoked         bool                        `json:"isRevoked" db:"isRevoked"`
	WorkflowId        string                      `json:"workflowId" db:"workflowRef"`
	WorkflowRunId     string                      `json:"workflowRunId" db:"workflowRunRef"`
	WorkflowNodeId    string                      `json:"workflowNodeId" db:"workflowNodeId"`
	DeletedAt         *time.Time                  `json:"deleted" db:"deleted"`
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
