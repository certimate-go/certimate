package dtos

import "time"

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
