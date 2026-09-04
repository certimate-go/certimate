package purgers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	pgrimpl "github.com/certimate-go/certimate/pkg/core/purger/providers/azure-keyvault"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.PurgeProviderTypeAzureKeyVault, func(options *ProviderFactoryOptions) (core.Purger, error) {
		credentials := domain.AccessConfigForAzure{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := pgrimpl.NewPurger(&pgrimpl.PurgerConfig{
			TenantId:     credentials.TenantId,
			ClientId:     credentials.ClientId,
			ClientSecret: credentials.ClientSecret,
			CloudName:    credentials.CloudName,
			KeyVaultName: xmaps.GetString(options.ProviderExtendedConfig, "keyvaultName"),
		})
		return provider, err
	})
}
