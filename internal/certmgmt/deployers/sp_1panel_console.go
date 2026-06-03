package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/1panel-console"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderType1PanelConsole, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigFor1Panel{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			ServerUrl:                credentials.ServerUrl,
			ApiVersion:               credentials.ApiVersion,
			ApiKey:                   credentials.ApiKey,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
			AutoRestart:              xmaps.GetBool(options.ProviderExtendedConfig, "autoRestart"),
		})
		return provider, err
	})
}
