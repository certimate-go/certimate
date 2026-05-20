package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	uclouduclb "github.com/certimate-go/certimate/pkg/core/deployer/providers/ucloud-uclb"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeUCloudUCLB, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForUCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := uclouduclb.NewDeployer(&uclouduclb.DeployerConfig{
			PrivateKey:     credentials.PrivateKey,
			PublicKey:      credentials.PublicKey,
			ProjectId:      credentials.ProjectId,
			Region:         xmaps.GetString(options.ProviderExtendedConfig, "region"),
			DeployTarget:   xmaps.GetString(options.ProviderExtendedConfig, "deployTarget"),
			LoadbalancerId: xmaps.GetString(options.ProviderExtendedConfig, "loadbalancerId"),
			VServerId:      xmaps.GetString(options.ProviderExtendedConfig, "vserverId"),
		})
		return provider, err
	})
}
