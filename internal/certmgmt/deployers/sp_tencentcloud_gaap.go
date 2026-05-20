package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	tencentcloudgaap "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-gaap"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeTencentCloudGAAP, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForTencentCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := tencentcloudgaap.NewDeployer(&tencentcloudgaap.DeployerConfig{
			SecretId:     credentials.SecretId,
			SecretKey:    credentials.SecretKey,
			Endpoint:     xmaps.GetString(options.ProviderExtendedConfig, "endpoint"),
			DeployTarget: xmaps.GetString(options.ProviderExtendedConfig, "deployTarget"),
			ProxyId:      xmaps.GetString(options.ProviderExtendedConfig, "proxyId"),
			ListenerId:   xmaps.GetString(options.ProviderExtendedConfig, "listenerId"),
		})
		return provider, err
	})
}
