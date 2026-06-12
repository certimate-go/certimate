package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/linode-los"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeLinodeLOS, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForLinode{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			AccessToken: credentials.AccessToken,
			RegionId:    xmaps.GetString(options.ProviderExtendedConfig, "regionId"),
			Bucket:      xmaps.GetString(options.ProviderExtendedConfig, "bucket"),
		})
		return provider, err
	})
}
