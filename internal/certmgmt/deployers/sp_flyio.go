package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	flyio "github.com/certimate-go/certimate/pkg/core/deployer/providers/flyio"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeFlyIO, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForFlyIO{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := flyio.NewDeployer(&flyio.DeployerConfig{
			ApiToken: credentials.ApiToken,
			AppName:  xmaps.GetString(options.ProviderExtendedConfig, "appName"),
			Domain:   xmaps.GetString(options.ProviderExtendedConfig, "domain"),
		})
		return provider, err
	})
}
