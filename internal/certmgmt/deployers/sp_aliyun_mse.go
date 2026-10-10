package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-mse"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeAliyunMSE, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForAliyun{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}
		return dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			AccessKeyId:        credentials.AccessKeyId,
			AccessKeySecret:    credentials.AccessKeySecret,
			ResourceGroupId:    credentials.ResourceGroupId,
			Region:             xmaps.GetString(options.ProviderExtendedConfig, "region"),
			GatewayId:          xmaps.GetString(options.ProviderExtendedConfig, "gatewayId"),
			DomainMatchPattern: xmaps.GetString(options.ProviderExtendedConfig, "domainMatchPattern"),
			Domain:             xmaps.GetString(options.ProviderExtendedConfig, "domain"),
			ForceHttps:         xmaps.GetBool(options.ProviderExtendedConfig, "forceHttps"),
		})
	})
}
