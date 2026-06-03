package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/dokploy"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeDokploy, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForDokploy{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			ServerUrl:                credentials.ServerUrl,
			ApiKey:                   credentials.ApiKey,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
		})
		return provider, err
	})
}
