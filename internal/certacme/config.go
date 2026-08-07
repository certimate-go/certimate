package certacme

import (
	"context"
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/settings"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

type ACMEConfigOptions struct {
	CAProvider               domain.CAProviderType
	CAProviderAccessConfig   map[string]any
	CAProviderExtendedConfig map[string]any
	CertifierKeyAlgorithm    domain.CertificateKeyAlgorithmType
}

type ACMEConfig struct {
	CAProvider domain.CAProviderType
	CADirUrl   string
	EABKid     string
	EABHmacKey string
}

func CreateACMEConfig(ctx context.Context, options *ACMEConfigOptions) (*ACMEConfig, error) {
	if options == nil {
		return nil, fmt.Errorf("the options is nil")
	}

	provider := options.CAProvider
	providerAccessCfg := options.CAProviderAccessConfig

	if provider.String() == "" {
		globalSettingsForSSLProvider := settings.GetGlobalSettingsForSSLProvider()
		provider = globalSettingsForSSLProvider.Provider
		providerAccessCfg = globalSettingsForSSLProvider.Configs[globalSettingsForSSLProvider.Provider]
	}

	if provider.String() == "" {
		// default CA: Let's Encrypt
		provider = domain.CAProviderTypeLetsEncrypt
	}

	acmeDirUrl, err := getCADirUrl(provider, providerAccessCfg, options.CertifierKeyAlgorithm)
	if err != nil {
		return nil, err
	}

	acmeEab := domain.AccessConfigForACMEExternalAccountBinding{}
	if err := xmaps.Populate(providerAccessCfg, &acmeEab); err != nil {
		return nil, err
	}

	return &ACMEConfig{
		CAProvider: provider,
		CADirUrl:   acmeDirUrl,
		EABKid:     acmeEab.EabKid,
		EABHmacKey: acmeEab.EabHmacKey,
	}, nil
}
