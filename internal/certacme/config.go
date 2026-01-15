package certacme

import (
	"context"
	"errors"
	"strings"

	"github.com/go-acme/lego/v4/certcrypto"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/repository"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

var acmeDirUrls = map[string]string{
	string(domain.CAProviderTypeLetsEncrypt):         "https://acme-v02.api.letsencrypt.org/directory",
	string(domain.CAProviderTypeLetsEncryptStaging):  "https://acme-staging-v02.api.letsencrypt.org/directory",
	string(domain.CAProviderTypeActalisSSL):          "https://acme-api.actalis.com/acme/directory",
	string(domain.CAProviderTypeGlobalSignAtlas):     "https://emea.acme.atlas.globalsign.com/directory",
	string(domain.CAProviderTypeGoogleTrustServices): "https://dv.acme-v02.api.pki.goog/directory",
	string(domain.CAProviderTypeLiteSSL):             "https://acme.litessl.com/acme/v2/directory",
	string(domain.CAProviderTypeSSLCom):              "https://acme.ssl.com/sslcom-dv-rsa",
	string(domain.CAProviderTypeSSLCom) + "RSA":      "https://acme.ssl.com/sslcom-dv-rsa",
	string(domain.CAProviderTypeSSLCom) + "ECC":      "https://acme.ssl.com/sslcom-dv-ecc",
	string(domain.CAProviderTypeSectigo):             "https://acme.sectigo.com/v2/DV",
	string(domain.CAProviderTypeSectigo) + "DV":      "https://acme.sectigo.com/v2/DV",
	string(domain.CAProviderTypeSectigo) + "OV":      "https://acme.sectigo.com/v2/OV",
	string(domain.CAProviderTypeSectigo) + "EV":      "https://acme.sectigo.com/v2/EV",
	string(domain.CAProviderTypeZeroSSL):             "https://acme.zerossl.com/v2/DV90",
}

type ACMEConfigOptions struct {
	CAProvider       string
	CAAccessConfig   map[string]any
	CAProviderConfig map[string]any
	CertifierKeyType certcrypto.KeyType
}

type ACMEConfig struct {
	CAProvider       domain.CAProviderType
	CADirUrl         string
	EABKid           string
	EABHmacKey       string
	CertifierKeyType certcrypto.KeyType
}

func NewACMEConfig(options *ACMEConfigOptions) (*ACMEConfig, error) {
	if options == nil {
		return nil, errors.New("the options is nil")
	}

	caProvider := options.CAProvider
	caAccessConfig := options.CAAccessConfig

	if options.CAProvider == "" {
		settingsRepo := repository.NewSettingsRepository()
		settings, _ := settingsRepo.GetByName(context.Background(), domain.SettingsNameSSLProvider)
		if settings != nil {
			sslProviderSettings := settings.Content.AsSSLProvider()
			caProvider = string(sslProviderSettings.Provider)
			caAccessConfig = sslProviderSettings.Configs[sslProviderSettings.Provider]
		}
	}

	if caProvider == "" {
		// default CA: Let's Encrypt
		caProvider = string(domain.AccessProviderTypeLetsEncrypt)
	}

	if caAccessConfig == nil {
		caAccessConfig = make(map[string]any)
	}

	ca := &ACMEConfig{CAProvider: domain.CAProviderType(caProvider), CertifierKeyType: options.CertifierKeyType}
	switch ca.CAProvider {
	case domain.CAProviderTypeSectigo:
		credentials := &domain.AccessConfigForGlobalSectigo{}
		if err := xmaps.Populate(caAccessConfig, &credentials); err != nil {
			return nil, err
		} else if strings.EqualFold(credentials.ValidationType, "DV") {
			ca.CADirUrl = acmeDirUrls[string(domain.CAProviderTypeSectigo)+"DV"]
		} else if strings.EqualFold(credentials.ValidationType, "OV") {
			ca.CADirUrl = acmeDirUrls[string(domain.CAProviderTypeSectigo)+"OV"]
		} else if strings.EqualFold(credentials.ValidationType, "EV") {
			ca.CADirUrl = acmeDirUrls[string(domain.CAProviderTypeSectigo)+"EV"]
		} else {
			ca.CADirUrl = acmeDirUrls[string(domain.CAProviderTypeSectigo)]
		}

	case domain.CAProviderTypeSSLCom:
		if strings.HasPrefix(string(options.CertifierKeyType), "RSA") {
			ca.CADirUrl = acmeDirUrls[string(domain.CAProviderTypeSSLCom)+"RSA"]
		} else if strings.HasPrefix(string(options.CertifierKeyType), "EC") {
			ca.CADirUrl = acmeDirUrls[string(domain.CAProviderTypeSSLCom)+"ECC"]
		} else {
			ca.CADirUrl = acmeDirUrls[string(domain.CAProviderTypeSSLCom)]
		}

	case domain.CAProviderTypeACMECA:
		credentials := &domain.AccessConfigForACMECA{}
		if err := xmaps.Populate(caAccessConfig, &credentials); err != nil {
			return nil, err
		} else if credentials.Endpoint == "" {
			return nil, errors.New("the endpoint of custom ACME CA is empty")
		}
		ca.CADirUrl = credentials.Endpoint

	default:
		endpoint := acmeDirUrls[string(ca.CAProvider)]
		if endpoint == "" {
			return nil, errors.New("the endpoint of the ACME CA provider is empty")
		}
		ca.CADirUrl = endpoint
	}

	eab := domain.AccessConfigForACMEExternalAccountBinding{}
	if err := xmaps.Populate(caAccessConfig, &eab); err != nil {
		return nil, err
	}
	ca.EABKid = eab.EabKid
	ca.EABHmacKey = eab.EabHmacKey

	return ca, nil
}
