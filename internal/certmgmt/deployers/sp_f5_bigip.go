package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/f5bigip"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeF5BigIP, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForF5BigIP{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			ServerUrl:                credentials.ServerUrl,
			Username:                 credentials.Username,
			Password:                 credentials.Password,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
			DeployTarget:             xmaps.GetString(options.ProviderExtendedConfig, "deployTarget"),
			Partition:                xmaps.GetString(options.ProviderExtendedConfig, "partition"),
			ClientSSLProfileName:     xmaps.GetString(options.ProviderExtendedConfig, "clientSSLProfileName"),
		})
		return provider, err
	})
}
