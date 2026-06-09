package domain

import (
	"crypto/x509"
	"fmt"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/certcrypto"

	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xcertkey "github.com/certimate-go/certimate/pkg/utils/cert/key"
	xcertx509 "github.com/certimate-go/certimate/pkg/utils/cert/x509"
)

const CollectionNameCertificate = "certificate"

type Certificate struct {
	Meta
	Source            CertificateSourceType           `db:"source"            json:"source"`
	Certificate       string                          `db:"certificate"       json:"certificate"`
	PrivateKey        string                          `db:"privateKey"        json:"privateKey"`
	SerialNumber      string                          `db:"serialNumber"      json:"serialNumber"`
	SubjectName       string                          `db:"subjectName"       json:"subjectName"`
	SubjectAltNames   string                          `db:"subjectAltNames"   json:"subjectAltNames"`
	IssuerName        string                          `db:"issuerName"        json:"issuerName"`
	IssuerOrg         string                          `db:"issuerOrg"         json:"issuerOrg"`
	IssuerCertificate string                          `db:"issuerCertificate" json:"issuerCertificate"`
	KeyAlgorithm      CertificateKeyAlgorithmType     `db:"keyAlgorithm"      json:"keyAlgorithm"`
	ValidationPolicy  CertificateValidationPolicyType `db:"validationPolicy"  json:"validationPolicy"`
	ValidityNotBefore time.Time                       `db:"validityNotBefore" json:"validityNotBefore"`
	ValidityNotAfter  time.Time                       `db:"validityNotAfter"  json:"validityNotAfter"`
	ValidityInterval  int32                           `db:"validityInterval"  json:"validityInterval"`
	ACMEAcctUrl       string                          `db:"acmeAcctUrl"       json:"acmeAcctUrl"`
	ACMECertUrl       string                          `db:"acmeCertUrl"       json:"acmeCertUrl"`
	IsRenewed         bool                            `db:"isRenewed"         json:"isRenewed"`
	IsRevoked         bool                            `db:"isRevoked"         json:"isRevoked"`
	WorkflowId        string                          `db:"workflowRef"       json:"workflowId"`
	WorkflowRunId     string                          `db:"workflowRunRef"    json:"workflowRunId"`
	WorkflowNodeId    string                          `db:"workflowNodeId"    json:"workflowNodeId"`
	DeletedAt         *time.Time                      `db:"deleted" json:"deleted"`
}

func (c *Certificate) PopulateFromX509(certX509 *x509.Certificate) *Certificate {
	c.SerialNumber = strings.ToUpper(certX509.SerialNumber.Text(16))
	c.SubjectName = certX509.Subject.CommonName
	c.SubjectAltNames = strings.Join(xcertx509.GetSubjectAltNames(certX509), ";")
	c.IssuerName = certX509.Issuer.CommonName
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

	validationType := xcertx509.GetValidationType(certX509)
	switch validationType {
	case xcertx509.ExtendedValidation:
		c.ValidationPolicy = CertificateValidationPolicyTypeEV
	case xcertx509.DomainValidated:
		c.ValidationPolicy = CertificateValidationPolicyTypeDV
	case xcertx509.OrganizationalValidated:
		c.ValidationPolicy = CertificateValidationPolicyTypeOV
	case xcertx509.IndividualValidated:
		c.ValidationPolicy = CertificateValidationPolicyTypeIV
	default:
		c.ValidationPolicy = CertificateValidationPolicyType("")
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

func (t CertificateSourceType) String() string {
	return string(t)
}

const (
	CertificateSourceTypeRequest = CertificateSourceType("request")
	CertificateSourceTypeUpload  = CertificateSourceType("upload")
)

type CertificateKeyAlgorithmType certcrypto.KeyType

func (t CertificateKeyAlgorithmType) String() string {
	return string(t)
}

func (t CertificateKeyAlgorithmType) LegoKeyType() certcrypto.KeyType {
	return certcrypto.KeyType(t)
}

const (
	CertificateKeyAlgorithmTypeRSA2048 = CertificateKeyAlgorithmType(certcrypto.RSA2048)
	CertificateKeyAlgorithmTypeRSA3072 = CertificateKeyAlgorithmType(certcrypto.RSA3072)
	CertificateKeyAlgorithmTypeRSA4096 = CertificateKeyAlgorithmType(certcrypto.RSA4096)
	CertificateKeyAlgorithmTypeRSA8192 = CertificateKeyAlgorithmType(certcrypto.RSA8192)
	CertificateKeyAlgorithmTypeEC256   = CertificateKeyAlgorithmType(certcrypto.EC256)
	CertificateKeyAlgorithmTypeEC384   = CertificateKeyAlgorithmType(certcrypto.EC384)
)

type CertificateValidationPolicyType string

func (t CertificateValidationPolicyType) String() string {
	return string(t)
}

const (
	CertificateValidationPolicyTypeEV = CertificateValidationPolicyType("EV")
	CertificateValidationPolicyTypeDV = CertificateValidationPolicyType("DV")
	CertificateValidationPolicyTypeOV = CertificateValidationPolicyType("OV")
	CertificateValidationPolicyTypeIV = CertificateValidationPolicyType("IV")
)

type CertificateFormatType string

func (t CertificateFormatType) String() string {
	return string(t)
}

const (
	CertificateFormatTypePEM CertificateFormatType = "PEM"
	CertificateFormatTypePFX CertificateFormatType = "PFX"
	CertificateFormatTypeJKS CertificateFormatType = "JKS"
)
