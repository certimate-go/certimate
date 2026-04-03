package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	huaweicloudaad "github.com/certimate-go/certimate/pkg/core/deployer/providers/huaweicloud-aad"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeHuaweiCloudAAD, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForHuaweiCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := huaweicloudaad.NewDeployer(&huaweicloudaad.DeployerConfig{
			AccessKeyId:         credentials.AccessKeyId,
			SecretAccessKey:     credentials.SecretAccessKey,
			EnterpriseProjectId: credentials.EnterpriseProjectId,
			InstanceId:          xmaps.GetString(options.ProviderExtendedConfig, "instanceId"),
			DomainMatchPattern:  xmaps.GetString(options.ProviderExtendedConfig, "domainMatchPattern"),
			Domain:              xmaps.GetString(options.ProviderExtendedConfig, "domain"),
		})
		return provider, err
	})
}
