package certacme

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/lego"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/repository"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

func RotateACMEAccountKey(ctx context.Context, accountId string) (*ACMEAccount, error) {
	if accountId == "" {
		return nil, domain.NewError(400, "the account id is empty")
	}

	accountRepo := repository.NewACMEAccountRepository()
	account, err := accountRepo.GetById(ctx, accountId)
	if err != nil {
		if domain.IsRecordNotFoundError(err) {
			return nil, domain.ErrRecordNotFound
		}
		return nil, fmt.Errorf("failed to get acme account: %w", err)
	}

	currentKey := account.GetPrivateKey()
	if currentKey == nil {
		return nil, domain.NewError(400, "stored private key is invalid or empty")
	}

	newKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new account key: %w", err)
	}
	newKeyPEM, err := xcert.ConvertECPrivateKeyToPEM(newKey, false)
	if err != nil {
		return nil, fmt.Errorf("failed to encode new account key: %w", err)
	}

	reg := account.GetRegistration()
	if reg == nil || reg.Location == "" {
		return nil, domain.NewError(400, "account registration location is empty")
	}

	legoUser := &importLegoUser{
		email:        account.Email,
		privateKey:   currentKey,
		registration: reg,
	}

	legoCfg := lego.NewConfig(legoUser)
	legoCfg.UserAgent = app.AppUserAgent
	legoCfg.CADirURL = account.ACMEDirectoryUrl
	legoClient, err := lego.NewClient(legoCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create lego client: %w", err)
	}

	if err := legoClient.Registration.KeyRollover(ctx, newKey); err != nil {
		return nil, fmt.Errorf("failed to rollover acme account key: %w", err)
	}

	legoUser.privateKey = newKey
	legoUser.registration = &acme.ExtendedAccount{
		Account:  *account.ResourceObject,
		Location: account.ACMEAccountUrl,
	}
	legoCfg2 := lego.NewConfig(legoUser)
	legoCfg2.UserAgent = app.AppUserAgent
	legoCfg2.CADirURL = account.ACMEDirectoryUrl
	legoClient2, err := lego.NewClient(legoCfg2)
	if err == nil {
		if refreshed, qerr := legoClient2.Registration.QueryRegistration(ctx); qerr == nil && refreshed != nil {
			account.ResourceObject = &refreshed.Account
		}
	}

	account.PrivateKey = newKeyPEM
	if _, err := accountRepo.Save(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to save rotated acme account: %w", err)
	}

	return account, nil
}
