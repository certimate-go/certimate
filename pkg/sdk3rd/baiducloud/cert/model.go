package cert

import (
	"github.com/baidubce/bce-sdk-go/services/cert"
)

type CreateCertArgs struct {
	cert.CreateCertArgs
}

type CreateCertResult struct {
	cert.CreateCertResult
}

type UpdateCertNameArgs struct {
	cert.UpdateCertNameArgs
}

type CertificateMeta struct {
	cert.CertificateMeta
}

type CertificateDetailMeta struct {
	cert.CertificateDetailMeta
}

type CertificateRawData struct {
	CertId          string `json:"certId"`
	CertName        string `json:"certName"`
	CertServerData  string `json:"certServerData"`
	CertPrivateData string `json:"certPrivateKey"`
	CertLinkData    string `json:"certLinkData,omitempty"`
	CertType        int    `json:"certType,omitempty"`
}

type ListCertResult struct {
	cert.ListCertResult
}

type ListCertDetailResult struct {
	cert.ListCertDetailResult
}

type UpdateCertDataArgs struct {
	cert.UpdateCertDataArgs
}

type CertInServiceMeta struct {
	cert.CertInServiceMeta
}
