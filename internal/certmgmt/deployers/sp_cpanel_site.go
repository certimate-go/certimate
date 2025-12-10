package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	opsite "github.com/certimate-go/certimate/pkg/core/deployer/providers/cpanel-site"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeCPanelSite, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForCPanel{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := opsite.NewDeployer(&opsite.DeployerConfig{
			ServerUrl:                credentials.ServerUrl,
			Username:                 credentials.Username,
			ApiToken:                 credentials.ApiToken,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
			Domain:                   xmaps.GetString(options.ProviderExtendedConfig, "domain"),
		})
		return provider, err
	})
}
