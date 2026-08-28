package certacme

import (
	"context"
	"crypto"
	"fmt"
	"strings"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/lego"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/repository"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type ImportACMEAccountOptions struct {
	CA            string
	ACMEDirUrl    string
	PrivateKeyPem string
	Email         string
}

func ImportACMEAccount(ctx context.Context, opts *ImportACMEAccountOptions) (*ACMEAccount, error) {
	if opts == nil {
		return nil, domain.NewError(400, "the import options are nil")
	}
	if strings.TrimSpace(opts.CA) == "" {
		return nil, domain.NewError(400, "the ca is empty")
	}
	if strings.TrimSpace(opts.PrivateKeyPem) == "" {
		return nil, domain.NewError(400, "the private key is empty")
	}

	dirUrl, err := resolveImportDirectoryURL(opts.CA, opts.ACMEDirUrl)
	if err != nil {
		return nil, err
	}

	privKey, err := xcert.ParsePrivateKeyFromPEM(opts.PrivateKeyPem)
	if err != nil {
		return nil, domain.NewError(400, fmt.Sprintf("invalid private key pem: %v", err))
	}
	signer, ok := privKey.(crypto.Signer)
	if !ok {
		return nil, domain.NewError(400, "private key is not a crypto.Signer")
	}

	keyPEM, err := xcert.ConvertPrivateKeyToPEM(privKey, false)
	if err != nil {
		keyPEM = opts.PrivateKeyPem
	}

	placeholder := &acme.Account{}
	tempEmail := strings.TrimSpace(opts.Email)
	if tempEmail == "" {
		tempEmail = "import@local"
	}

	legoUser := &importLegoUser{
		email:      tempEmail,
		privateKey: signer,
		registration: &acme.ExtendedAccount{
			Account:  *placeholder,
			Location: "",
		},
	}

	legoCfg := lego.NewConfig(legoUser)
	legoCfg.UserAgent = app.AppUserAgent
	legoCfg.CADirURL = dirUrl
	legoClient, err := lego.NewClient(legoCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create lego client: %w", err)
	}

	resolved, err := legoClient.Registration.ResolveAccountByKey(ctx)
	if err != nil {
		return nil, domain.NewError(404, fmt.Sprintf("failed to resolve acme account by key: %v", err))
	}
	if resolved == nil || resolved.Location == "" {
		return nil, domain.NewError(404, "acme account not found for the given private key")
	}

	email := extractEmailFromContacts(resolved.Contact)
	if email == "" {
		email = strings.TrimSpace(opts.Email)
	}
	if email == "" {
		return nil, domain.NewError(400, "the account has no email contact; please provide an email")
	}

	accountRepo := repository.NewACMEAccountRepository()
	existing, err := accountRepo.GetByCAAndAcctUrl(ctx, opts.CA, resolved.Location)
	if err != nil && !domain.IsRecordNotFoundError(err) {
		return nil, fmt.Errorf("failed to check existing acme account: %w", err)
	}
	if existing != nil {
		return nil, domain.NewError(409, "acme account already exists")
	}

	if byEmail, err := accountRepo.GetByCAAndEmail(ctx, opts.CA, dirUrl, email); err == nil && byEmail != nil {
		if byEmail.ACMEAccountUrl == resolved.Location {
			return nil, domain.NewError(409, "acme account already exists")
		}
	} else if err != nil && !domain.IsRecordNotFoundError(err) {
		return nil, fmt.Errorf("failed to check existing acme account by email: %w", err)
	}

	account := &ACMEAccount{
		CA:               opts.CA,
		Email:            email,
		PrivateKey:       keyPEM,
		ACMEDirectoryUrl: dirUrl,
		ACMEAccountUrl:   resolved.Location,
		ResourceObject:   &resolved.Account,
	}

	if _, err := accountRepo.Save(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to save acme account record: %w", err)
	}

	return account, nil
}

func resolveImportDirectoryURL(ca, explicitDir string) (string, error) {
	if strings.TrimSpace(explicitDir) != "" {
		return strings.TrimSpace(explicitDir), nil
	}

	provider := domain.CAProviderType(ca)
	if provider == domain.CAProviderTypeACMECA {
		return "", domain.NewError(400, "acmeDirUrl is required for custom ACME CA")
	}

	dirUrl, err := getCADirUrl(provider, nil, "")
	if err != nil {
		return "", domain.NewError(400, err.Error())
	}
	return dirUrl, nil
}

func extractEmailFromContacts(contacts []string) string {
	for _, c := range contacts {
		c = strings.TrimSpace(c)
		if strings.HasPrefix(strings.ToLower(c), "mailto:") {
			email := strings.TrimSpace(c[len("mailto:"):])
			if email != "" {
				return email
			}
		}
	}
	return ""
}

type importLegoUser struct {
	email        string
	privateKey   crypto.Signer
	registration *acme.ExtendedAccount
}

func (u *importLegoUser) GetEmail() string {
	return u.email
}

func (u *importLegoUser) GetRegistration() *acme.ExtendedAccount {
	return u.registration
}

func (u *importLegoUser) GetPrivateKey() crypto.Signer {
	return u.privateKey
}
