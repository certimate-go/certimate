package certacme

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net/mail"
	"strings"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/registration"
	"golang.org/x/sync/singleflight"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/repository"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

var registrationSg singleflight.Group

type ACMEAccount = domain.ACMEAccount

type acmeAccountRepository interface {
	GetByCAAndEmail(ctx context.Context, ca, caDirUrl, email string) (*ACMEAccount, error)
	Save(ctx context.Context, acmeAccount *ACMEAccount) (*ACMEAccount, error)
}

type acmeRegistrationClient interface {
	ExternalAccountRequired() bool
	Register(ctx context.Context, eab *eabCredential) (*acme.ExtendedAccount, error)
}

type acmeRegistrationClientFactory func(config *ACMEConfig, account *ACMEAccount) (acmeRegistrationClient, error)

type acmeRegistrationDeps struct {
	accountRepo     acmeAccountRepository
	clientFactory   acmeRegistrationClientFactory
	findEABProvider func(ca domain.CAProviderType) ACMEEABProvider
}

func defaultACMERegistrationDeps() *acmeRegistrationDeps {
	return &acmeRegistrationDeps{
		accountRepo:     repository.NewACMEAccountRepository(),
		clientFactory:   newLegoRegistrationClient,
		findEABProvider: findACMEEABProvider,
	}
}

type legoRegistrationClient struct {
	client      *lego.Client
	eabRequired bool
}

func newLegoRegistrationClient(config *ACMEConfig, account *ACMEAccount) (acmeRegistrationClient, error) {
	legoCfg := lego.NewConfig(account)
	legoCfg.UserAgent = app.AppUserAgent
	legoCfg.CADirURL = config.CADirUrl

	legoClient, err := lego.NewClient(legoCfg)
	if err != nil {
		return nil, err
	}

	return &legoRegistrationClient{
		client:      legoClient,
		eabRequired: legoClient.GetServerMetadata().ExternalAccountRequired,
	}, nil
}

func (c *legoRegistrationClient) ExternalAccountRequired() bool {
	return c.eabRequired
}

func (c *legoRegistrationClient) Register(ctx context.Context, eab *eabCredential) (*acme.ExtendedAccount, error) {
	if eab != nil {
		return c.client.Registration.RegisterWithExternalAccountBinding(ctx, registration.RegisterEABOptions{
			TermsOfServiceAgreed: true,
			Kid:                  eab.Kid,
			HmacEncoded:          eab.HmacKeyEncoded,
		})
	}

	return c.client.Registration.Register(ctx, registration.RegisterOptions{
		TermsOfServiceAgreed: true,
	})
}

func CreateACMEAccount(ctx context.Context, config *ACMEConfig, email string) (*ACMEAccount, error) {
	return acquireACMEAccount(ctx, config, email, defaultACMERegistrationDeps())
}

func CreateACMEAccountWithSingleFlight(ctx context.Context, config *ACMEConfig, email string) (*ACMEAccount, error) {
	return acquireACMEAccount(ctx, config, email, defaultACMERegistrationDeps())
}

func acquireACMEAccount(ctx context.Context, config *ACMEConfig, email string, deps *acmeRegistrationDeps) (*ACMEAccount, error) {
	if config == nil {
		return nil, fmt.Errorf("the acme config is nil")
	}
	if email == "" {
		return nil, fmt.Errorf("the email is empty")
	}

	key := fmt.Sprintf("%s|%s|%s", config.CAProvider, config.CADirUrl, email)
	resultCh := registrationSg.DoChan(key, func() (any, error) {
		return registerACMEAccount(ctx, deps, config, email)
	})

	select {
	case result := <-resultCh:
		if result.Err != nil {
			return nil, result.Err
		}
		return result.Val.(*ACMEAccount), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func registerACMEAccount(ctx context.Context, deps *acmeRegistrationDeps, config *ACMEConfig, email string) (*ACMEAccount, error) {
	account, err := deps.accountRepo.GetByCAAndEmail(ctx, config.CAProvider.String(), config.CADirUrl, email)
	if err != nil {
		if !domain.IsRecordNotFoundError(err) {
			return nil, fmt.Errorf("failed to get acme account record: %w", err)
		}
	}

	if account != nil {
		return account, nil
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	keyPEM, err := xcert.ConvertECPrivateKeyToPEM(key, false)
	if err != nil {
		return nil, err
	}

	account = &ACMEAccount{
		CA:               config.CAProvider.String(),
		Email:            email,
		PrivateKey:       keyPEM,
		ACMEDirectoryUrl: config.CADirUrl,
	}

	client, err := deps.clientFactory(config, account)
	if err != nil {
		return nil, err
	}

	var eab *eabCredential
	if client.ExternalAccountRequired() {
		eab, err = resolveEABCredentialsForRegistration(ctx, deps, config, email)
		if err != nil {
			return nil, err
		}
	}

	regres, regerr := client.Register(ctx, eab)
	if regerr != nil {
		return nil, fmt.Errorf("failed to register acme account: %w", regerr)
	}

	account.ACMEAccountUrl = regres.Location
	account.ResourceObject = &regres.Account

	if _, err := deps.accountRepo.Save(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to save acme account record: %w", err)
	}

	return account, nil
}

func resolveEABCredentialsForRegistration(ctx context.Context, deps *acmeRegistrationDeps, config *ACMEConfig, email string) (*eabCredential, error) {
	keyId := strings.TrimSpace(config.EABKid)
	hmacEncoded := normalizeEABHmacKeyEncoded(config.EABHmacKey)

	if keyId != "" && hmacEncoded != "" {
		return &eabCredential{Kid: keyId, HmacKeyEncoded: hmacEncoded}, nil
	}

	if keyId == "" && hmacEncoded == "" {
		provider := deps.findEABProvider(config.CAProvider)
		if provider == nil {
			return nil, fmt.Errorf("missing or invalid eab kid")
		}

		acquireEmail, err := normalizeEABAcquireEmail(email)
		if err != nil {
			return nil, err
		}

		credential, err := provider.AcquireEAB(ctx, acquireEmail)
		if err != nil {
			return nil, err
		}

		if provider.Reusable() {
			config.EABKid = credential.Kid
			config.EABHmacKey = credential.HmacKeyEncoded
		}
		return credential, nil
	}

	if keyId == "" {
		return nil, fmt.Errorf("missing or invalid eab kid")
	}

	return nil, fmt.Errorf("missing or invalid eab hmac key")
}

func normalizeEABAcquireEmail(email string) (string, error) {
	trimmed := strings.TrimSpace(email)
	if trimmed == "" {
		return "", fmt.Errorf("the registration email is empty")
	}

	parsed, err := mail.ParseAddress(trimmed)
	if err != nil || parsed.Address != trimmed {
		return "", fmt.Errorf("the registration email is invalid")
	}

	return trimmed, nil
}
