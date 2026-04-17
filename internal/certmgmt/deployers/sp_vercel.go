package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	vercel "github.com/certimate-go/certimate/pkg/core/deployer/providers/vercel"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeVercel, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForVercel{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := vercel.NewDeployer(&vercel.DeployerConfig{
			ApiAccessToken: credentials.ApiAccessToken,
			TeamId:         credentials.TeamId,
		})
		return provider, err
	})
}
