package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/cpanel"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeCPanel, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForCPanel{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			ServerUrl:                credentials.ServerUrl,
			Username:                 credentials.Username,
			ApiToken:                 credentials.ApiToken,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
			DeployTarget:             xmaps.GetString(options.ProviderExtendedConfig, "deployTarget"),
			Domain:                   xmaps.GetString(options.ProviderExtendedConfig, "domain"),
		})
		return provider, err
	})
}
