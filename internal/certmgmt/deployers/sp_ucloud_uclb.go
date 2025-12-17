package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	uclouduclb "github.com/certimate-go/certimate/pkg/core/deployer/providers/ucloud-uclb"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeUCloudUCLB, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForUCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := uclouduclb.NewDeployer(&uclouduclb.DeployerConfig{
			PrivateKey:     credentials.PrivateKey,
			PublicKey:      credentials.PublicKey,
			ProjectId:      credentials.ProjectId,
			Region:         xmaps.GetString(options.ProviderExtendedConfig, "region"),
			ResourceType:   xmaps.GetString(options.ProviderExtendedConfig, "resourceType"),
			LoadbalancerId: xmaps.GetString(options.ProviderExtendedConfig, "loadbalancerId"),
			VServerId:      xmaps.GetString(options.ProviderExtendedConfig, "vserverId"),
		})
		return provider, err
	})
}
