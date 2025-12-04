package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	mohuamvh "github.com/certimate-go/certimate/pkg/core/deployer/providers/mohua-mvh"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeMohuaMVH, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForMohua{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := mohuamvh.NewDeployer(&mohuamvh.DeployerConfig{
			Username:    credentials.Username,
			ApiPassword: credentials.ApiPassword,
			HostId:      xmaps.GetString(options.ProviderExtendedConfig, "hostId"),
			DomainId:    xmaps.GetString(options.ProviderExtendedConfig, "domainId"),
		})
		return provider, err
	})
}
