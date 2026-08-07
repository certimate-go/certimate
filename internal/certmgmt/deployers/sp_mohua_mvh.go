package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/mohua-mvh"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeMohuaMVH, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForMohua{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			Username:    credentials.Username,
			ApiPassword: credentials.ApiPassword,
			HostId:      xmaps.GetString(options.ProviderExtendedConfig, "hostId"),
			DomainId:    xmaps.GetInt64(options.ProviderExtendedConfig, "domainId"),
		})
		return provider, err
	})
}
