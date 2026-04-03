package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	huaweicloudapig "github.com/certimate-go/certimate/pkg/core/deployer/providers/huaweicloud-apig"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeHuaweiCloudAPIG, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForHuaweiCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := huaweicloudapig.NewDeployer(&huaweicloudapig.DeployerConfig{
			AccessKeyId:         credentials.AccessKeyId,
			SecretAccessKey:     credentials.SecretAccessKey,
			EnterpriseProjectId: credentials.EnterpriseProjectId,
			Region:              xmaps.GetString(options.ProviderExtendedConfig, "region"),
			ResourceType:        xmaps.GetString(options.ProviderExtendedConfig, "resourceType"),
			CertificateId:       xmaps.GetString(options.ProviderExtendedConfig, "certificateId"),
		})
		return provider, err
	})
}
