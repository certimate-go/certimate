package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	aliyunga "github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-ga"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeAliyunGA, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForAliyun{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := aliyunga.NewDeployer(&aliyunga.DeployerConfig{
			AccessKeyId:     credentials.AccessKeyId,
			AccessKeySecret: credentials.AccessKeySecret,
			ResourceGroupId: credentials.ResourceGroupId,
			DeployTarget:    xmaps.GetString(options.ProviderExtendedConfig, "deployTarget"),
			AcceleratorId:   xmaps.GetString(options.ProviderExtendedConfig, "acceleratorId"),
			ListenerId:      xmaps.GetString(options.ProviderExtendedConfig, "listenerId"),
			Domain:          xmaps.GetString(options.ProviderExtendedConfig, "domain"),
		})
		return provider, err
	})
}
