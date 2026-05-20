package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	aliyunwaf "github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-waf"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeAliyunWAF, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForAliyun{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := aliyunwaf.NewDeployer(&aliyunwaf.DeployerConfig{
			AccessKeyId:     credentials.AccessKeyId,
			AccessKeySecret: credentials.AccessKeySecret,
			ResourceGroupId: credentials.ResourceGroupId,
			Region:          xmaps.GetString(options.ProviderExtendedConfig, "region"),
			ServiceVersion:  xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "serviceVersion", "3.0"),
			ServiceType:     xmaps.GetString(options.ProviderExtendedConfig, "serviceType"),
			InstanceId:      xmaps.GetString(options.ProviderExtendedConfig, "instanceId"),
			ResourceProduct: xmaps.GetString(options.ProviderExtendedConfig, "resourceProduct"),
			ResourceId:      xmaps.GetString(options.ProviderExtendedConfig, "resourceId"),
			ResourcePort:    xmaps.GetOrDefaultInt32(options.ProviderExtendedConfig, "resourcePort", 443),
			Domain:          xmaps.GetString(options.ProviderExtendedConfig, "domain"),
		})
		return provider, err
	})
}
