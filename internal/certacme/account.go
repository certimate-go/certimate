package certacme

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"golang.org/x/sync/singleflight"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/repository"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

var registrationSg singleflight.Group

type ACMEAccount = domain.ACMEAccount

func NewACMEAccount(config *ACMEConfig, email string, register bool) (*ACMEAccount, error) {
	if config == nil {
		return nil, fmt.Errorf("the acme config is nil")
	}
	if email == "" {
		return nil, fmt.Errorf("the email is empty")
	}

	ctx := context.Background()
	accountRepo := repository.NewACMEAccountRepository()
	account, err := accountRepo.GetByCAAndEmail(ctx, config.CAProvider.String(), config.CADirUrl, email)
	if err != nil {
		if !domain.IsRecordNotFoundError(err) {
			return nil, fmt.Errorf("failed to get acme account record: %w", err)
		}
	}

	// register new acme account if not exists
	if account == nil {
		if !register {
			return nil, fmt.Errorf("the acme account does not exist")
		}

		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, err
		}

		keyPEM, err := xcert.ConvertECPrivateKeyToPEM(key)
		if err != nil {
			return nil, err
		}

		account = &ACMEAccount{
			CA:         config.CAProvider.String(),
			Email:      email,
			PrivateKey: keyPEM,
			ACMEDirUrl: config.CADirUrl,
		}
		legoCfg := lego.NewConfig(account)
		legoCfg.CADirURL = config.CADirUrl
		legoClient, err := lego.NewClient(legoCfg)
		if err != nil {
			return nil, err
		}

		var regres *registration.Resource
		var regerr error
		if legoClient.GetExternalAccountRequired() {
			if config.EABKid == "" {
				return nil, fmt.Errorf("missing or invalid eab kid")
			}
			if config.EABHmacKey == "" {
				return nil, fmt.Errorf("missing or invalid eab hmac key")
			}

			// patch, see https://github.com/go-acme/lego/issues/2634
			keyId := strings.TrimSpace(config.EABKid)
			keyEncoded := strings.TrimSpace(config.EABHmacKey)
			keyEncoded = strings.ReplaceAll(strings.ReplaceAll(keyEncoded, "+", "-"), "/", "_")
			keyEncoded = strings.TrimSuffix(keyEncoded, "=")

			regres, regerr = legoClient.Registration.RegisterWithExternalAccountBinding(registration.RegisterEABOptions{
				TermsOfServiceAgreed: true,
				Kid:                  keyId,
				HmacEncoded:          keyEncoded,
			})
		} else {
			regres, regerr = legoClient.Registration.Register(registration.RegisterOptions{
				TermsOfServiceAgreed: true,
			})
		}
		if regerr != nil {
			return nil, fmt.Errorf("failed to register acme account: %w", regerr)
		}

		account.ACMEAccount = &regres.Body
		account.ACMEAcctUrl = regres.URI

		if _, err := accountRepo.Save(ctx, account); err != nil {
			return nil, fmt.Errorf("failed to save acme account record: %w", err)
		}
	}

	return account, nil
}

func NewACMEAccountWithSingleFlight(config *ACMEConfig, email string) (*ACMEAccount, error) {
	if config == nil {
		return nil, fmt.Errorf("the acme config is nil")
	}
	if email == "" {
		return nil, fmt.Errorf("the email is empty")
	}

	resp, err, _ := registrationSg.Do(fmt.Sprintf("%s|%s|%s", config.CAProvider, config.CADirUrl, email), func() (any, error) {
		return NewACMEAccount(config, email, true)
	})
	if err != nil {
		return nil, err
	}

	return resp.(*ACMEAccount), nil
}
