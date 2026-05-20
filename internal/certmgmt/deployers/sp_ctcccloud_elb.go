package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	ctcccloudelb "github.com/certimate-go/certimate/pkg/core/deployer/providers/ctcccloud-elb"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeCTCCCloudELB, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForCTCCCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := ctcccloudelb.NewDeployer(&ctcccloudelb.DeployerConfig{
			AccessKeyId:     credentials.AccessKeyId,
			SecretAccessKey: credentials.SecretAccessKey,
			RegionId:        xmaps.GetString(options.ProviderExtendedConfig, "regionId"),
			ResourceType:    xmaps.GetString(options.ProviderExtendedConfig, "resourceType"),
			LoadbalancerId:  xmaps.GetString(options.ProviderExtendedConfig, "loadbalancerId"),
			ListenerId:      xmaps.GetString(options.ProviderExtendedConfig, "listenerId"),
		})
		return provider, err
	})
}
