package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/jdcloud-waf"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeJDCloudWAF, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForJDCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			AccessKeyId:     credentials.AccessKeyId,
			AccessKeySecret: credentials.AccessKeySecret,
			RegionId:        xmaps.GetString(options.ProviderExtendedConfig, "regionId"),
			InstanceId:      xmaps.GetString(options.ProviderExtendedConfig, "instanceId"),
			Domain:          xmaps.GetString(options.ProviderExtendedConfig, "domain"),
		})
		return provider, err
	})
}
