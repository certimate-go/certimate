package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	ucloudualb "github.com/certimate-go/certimate/pkg/core/deployer/providers/ucloud-ualb"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeUCloudUALB, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForUCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := ucloudualb.NewDeployer(&ucloudualb.DeployerConfig{
			PrivateKey:     credentials.PrivateKey,
			PublicKey:      credentials.PublicKey,
			ProjectId:      credentials.ProjectId,
			Region:         xmaps.GetString(options.ProviderExtendedConfig, "region"),
			ResourceType:   xmaps.GetString(options.ProviderExtendedConfig, "resourceType"),
			LoadbalancerId: xmaps.GetString(options.ProviderExtendedConfig, "loadbalancerId"),
			ListenerId:     xmaps.GetString(options.ProviderExtendedConfig, "listenerId"),
			Domain:         xmaps.GetString(options.ProviderExtendedConfig, "domain"),
		})
		return provider, err
	})
}
