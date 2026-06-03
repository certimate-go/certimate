package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-clb"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeAliyunCLB, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForAliyun{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			AccessKeyId:     credentials.AccessKeyId,
			AccessKeySecret: credentials.AccessKeySecret,
			ResourceGroupId: credentials.ResourceGroupId,
			Region:          xmaps.GetString(options.ProviderExtendedConfig, "region"),
			DeployTarget:    xmaps.GetString(options.ProviderExtendedConfig, "deployTarget"),
			LoadbalancerId:  xmaps.GetString(options.ProviderExtendedConfig, "loadbalancerId"),
			ListenerPort:    xmaps.GetOrDefaultInt32(options.ProviderExtendedConfig, "listenerPort", 443),
			Domain:          xmaps.GetString(options.ProviderExtendedConfig, "domain"),
		})
		return provider, err
	})
}
