package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-acme/lego/v5/acme"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/domain"
)

type ACMEAccountRepository struct{}

func NewACMEAccountRepository() *ACMEAccountRepository {
	return &ACMEAccountRepository{}
}

func (r *ACMEAccountRepository) GetByCAAndEmail(ctx context.Context, ca, caDirUrl, email string) (*domain.ACMEAccount, error) {
	record, err := app.GetApp().FindFirstRecordByFilter(
		domain.CollectionNameACMEAccount,
		"ca={:ca} && acmeDirUrl={:acmeDirUrl} && email={:email}",
		dbx.Params{"ca": ca, "acmeDirUrl": caDirUrl, "email": email},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	return r.castRecordToModel(record)
}

func (r *ACMEAccountRepository) GetByCAAndAcctUrl(ctx context.Context, ca string, acctUrl string) (*domain.ACMEAccount, error) {
	record, err := app.GetApp().FindFirstRecordByFilter(
		domain.CollectionNameACMEAccount,
		"ca={:ca} && acmeAcctUrl={:acmeAcctUrl}",
		dbx.Params{"ca": ca, "acmeAcctUrl": acctUrl},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	return r.castRecordToModel(record)
}

func (r *ACMEAccountRepository) Save(ctx context.Context, acmeAccount *domain.ACMEAccount) (*domain.ACMEAccount, error) {
	collection, err := app.GetApp().FindCollectionByNameOrId(domain.CollectionNameACMEAccount)
	if err != nil {
		return acmeAccount, err
	}

	var record *core.Record
	if acmeAccount.Id == "" {
		record = core.NewRecord(collection)
	} else {
		record, err = app.GetApp().FindRecordById(collection, acmeAccount.Id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return acmeAccount, domain.ErrRecordNotFound
			}
			return acmeAccount, err
		}
	}

	record.Set("ca", acmeAccount.CA)
	record.Set("email", acmeAccount.Email)
	record.Set("privateKey", acmeAccount.PrivateKey)
	record.Set("acmeDirUrl", acmeAccount.ACMEDirectoryUrl)
	record.Set("acmeAcctUrl", acmeAccount.ACMEAccountUrl)
	record.Set("resourceObj", acmeAccount.ResourceObject)
	if err := app.GetApp().Save(record); err != nil {
		return acmeAccount, err
	}

	acmeAccount.Id = record.Id
	acmeAccount.CreatedAt = record.GetDateTime("created").Time()
	acmeAccount.UpdatedAt = record.GetDateTime("updated").Time()
	return acmeAccount, nil
}

func (r *ACMEAccountRepository) castRecordToModel(record *core.Record) (*domain.ACMEAccount, error) {
	if record == nil {
		return nil, fmt.Errorf("the record is nil")
	}

	resourceObj := &acme.Account{}
	if err := record.UnmarshalJSONField("resourceObj", resourceObj); err != nil {
		return nil, fmt.Errorf("field 'resourceObj' is malformed")
	}

	acmeAccount := &domain.ACMEAccount{
		Meta: domain.Meta{
			Id:        record.Id,
			CreatedAt: record.GetDateTime("created").Time(),
			UpdatedAt: record.GetDateTime("updated").Time(),
		},
		CA:               record.GetString("ca"),
		Email:            record.GetString("email"),
		PrivateKey:       record.GetString("privateKey"),
		ACMEDirectoryUrl: record.GetString("acmeDirUrl"),
		ACMEAccountUrl:   record.GetString("acmeAcctUrl"),
		ResourceObject:   resourceObj,
	}
	return acmeAccount, nil
}
