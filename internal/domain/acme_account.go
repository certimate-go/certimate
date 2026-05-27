package domain

import (
	"crypto"

	"github.com/go-acme/lego/v5/acme"

	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

const CollectionNameACMEAccount = "acme_accounts"

type ACMEAccount struct {
	Meta
	CA          string        `db:"ca"          json:"ca"`
	Email       string        `db:"email"       json:"email"`
	PrivateKey  string        `db:"privateKey"  json:"privateKey"`
	ACMEDirUrl  string        `db:"acmeDirUrl"  json:"acmeDirUrl"`
	ACMEAcctUrl string        `db:"acmeAcctUrl" json:"acmeAcctUrl"`
	ACMEAccount *acme.Account `db:"acmeAccount" json:"acmeAccount"`
}

func (a *ACMEAccount) GetEmail() string {
	return a.Email
}

func (a *ACMEAccount) GetRegistration() *acme.ExtendedAccount {
	if a.ACMEAccount == nil {
		return nil
	}

	return &acme.ExtendedAccount{
		Account:  *a.ACMEAccount,
		Location: a.ACMEAcctUrl,
	}
}

func (a *ACMEAccount) GetPrivateKey() crypto.Signer {
	if a.PrivateKey == "" {
		return nil
	}

	rs, _ := xcert.ParsePrivateKeyFromPEM(a.PrivateKey)
	return rs.(crypto.Signer)
}
