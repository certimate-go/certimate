package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	aliyuncdn "github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-cdn"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeAliyunCDN, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForAliyun{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := aliyuncdn.NewDeployer(&aliyuncdn.DeployerConfig{
			AccessKeyId:        credentials.AccessKeyId,
			AccessKeySecret:    credentials.AccessKeySecret,
			ResourceGroupId:    credentials.ResourceGroupId,
			Region:             xmaps.GetString(options.ProviderExtendedConfig, "region"),
			DomainMatchPattern: xmaps.GetString(options.ProviderExtendedConfig, "domainMatchPattern"),
			Domain:             xmaps.GetString(options.ProviderExtendedConfig, "domain"),
		})
		return provider, err
	})
}
