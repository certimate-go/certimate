package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/azure-keyvault"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeAzureKeyVault, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForAzure{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			TenantId:        credentials.TenantId,
			ClientId:        credentials.ClientId,
			ClientSecret:    credentials.ClientSecret,
			CloudName:       credentials.CloudName,
			KeyVaultName:    xmaps.GetString(options.ProviderExtendedConfig, "keyvaultName"),
			CertificateName: xmaps.GetString(options.ProviderExtendedConfig, "certificateName"),
		})
		return provider, err
	})
}
