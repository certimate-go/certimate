package certacme

import (
	"context"
	"errors"
	"time"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/repository"
	"github.com/go-acme/lego/v4/lego"
)

type ACMEClient struct {
	client  *lego.Client
	account *ACMEAccount
}

func NewACMEClient(config *ACMEConfig, email string, configures ...func(*lego.Config) error) (*ACMEClient, error) {
	account, err := NewACMEAccountWithSingleFlight(config, email)
	if err != nil {
		return nil, err
	}

	mergedConfigures := []func(*lego.Config) error{
		func(legoCfg *lego.Config) error {
			legoCfg.CADirURL = config.CADirUrl
			legoCfg.Certificate.KeyType = config.CertifierKeyType
			return nil
		},
	}
	mergedConfigures = append(mergedConfigures, configures...)
	return newACMEClientWithAccount(account, mergedConfigures...)
}

func NewACMEClientWithAccount(account *ACMEAccount, configures ...func(*lego.Config) error) (*ACMEClient, error) {
	return newACMEClientWithAccount(account, configures...)
}

func newACMEClientWithAccount(account *ACMEAccount, configures ...func(*lego.Config) error) (*ACMEClient, error) {
	if account == nil {
		return nil, errors.New("the acme account is nil")
	}

	legoCfg := lego.NewConfig(account)
	legoCfg.CADirURL = account.ACMEDirUrl

	settingsRepo := repository.NewSettingsRepository()
	settings, _ := settingsRepo.GetByName(context.Background(), domain.SettingsNameSSLProvider)
	if settings != nil {
		sslProviderSettings := settings.Content.AsSSLProvider()
		if sslProviderSettings.Timeout > 0 {
			legoCfg.Certificate.Timeout = time.Duration(sslProviderSettings.Timeout) * time.Second
		}
	}

	errs := make([]error, 0)
	for _, configure := range configures {
		if err := configure(legoCfg); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	legoClient, err := lego.NewClient(legoCfg)
	if err != nil {
		return nil, err
	}

	return &ACMEClient{
		client:  legoClient,
		account: account,
	}, nil
}
