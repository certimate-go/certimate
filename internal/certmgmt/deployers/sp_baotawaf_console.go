package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	baotawafconsole "github.com/certimate-go/certimate/pkg/core/deployer/providers/baotawaf-console"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeBaotaWAFConsole, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForBaotaWAF{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := baotawafconsole.NewDeployer(&baotawafconsole.DeployerConfig{
			ServerUrl:                credentials.ServerUrl,
			ApiKey:                   credentials.ApiKey,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
		})
		return provider, err
	})
}
