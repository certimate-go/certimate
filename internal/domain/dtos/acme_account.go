package dtos

import (
	"time"

	"github.com/certimate-go/certimate/internal/domain"
)

type ACMEAccountView struct {
	Id               string    `json:"id"`
	CA               string    `json:"ca"`
	Email            string    `json:"email"`
	ACMEDirectoryUrl string    `json:"acmeDirUrl"`
	ACMEAccountUrl   string    `json:"acmeAcctUrl"`
	CreatedAt        time.Time `json:"created"`
	UpdatedAt        time.Time `json:"updated"`
}

type ACMEAccountListReq struct {
	Page    int    `json:"-"`
	PerPage int    `json:"-"`
	CA      string `json:"-"`
}

type ACMEAccountListResp struct {
	Items      []*ACMEAccountView `json:"items"`
	TotalItems int64              `json:"totalItems"`
	Page       int                `json:"page"`
	PerPage    int                `json:"perPage"`
}

type ACMEAccountImportReq struct {
	PrivateKeyPem string `json:"privateKeyPem"`
	CA            string `json:"ca"`
	ACMEDirUrl    string `json:"acmeDirUrl,omitempty"`
	Email         string `json:"email,omitempty"`
}

type ACMEAccountImportResp struct {
	Item *ACMEAccountView `json:"item"`
}

type ACMEAccountExportResp struct {
	PrivateKeyPem string `json:"privateKeyPem"`
}

type ACMEAccountRotateResp struct {
	Item *ACMEAccountView `json:"item"`
}

type DNSPersistPrepareReq struct {
	Domains            []string       `json:"domains"`
	ContactEmail       string         `json:"contactEmail"`
	CAProvider         string         `json:"caProvider"`
	CAProviderAccessId string         `json:"caProviderAccessId"`
	CAProviderConfig   map[string]any `json:"caProviderConfig"`
	KeyAlgorithm       string         `json:"keyAlgorithm"`
	ACMEProfile        string         `json:"acmeProfile"`
	Nameservers        []string       `json:"nameservers"`
}

type DNSPersistPrepareResp struct {
	AccountId  string                    `json:"accountId"`
	AccountUrl string                    `json:"accountUrl"`
	Records    []domain.DNSPersistRecord `json:"records"`
}

type DNSPersistCheckReq struct {
	Records     []domain.DNSPersistRecord `json:"records"`
	Nameservers []string                  `json:"nameservers"`
}

type DNSPersistCheckResp struct {
	Ready   bool                            `json:"ready"`
	Records []domain.DNSPersistRecordStatus `json:"records"`
}
