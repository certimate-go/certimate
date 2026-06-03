package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	tencentcloudvod "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-vod"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeTencentCloudVOD, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForTencentCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := tencentcloudvod.NewDeployer(&tencentcloudvod.DeployerConfig{
			SecretId:           credentials.SecretId,
			SecretKey:          credentials.SecretKey,
			ProjectId:          credentials.ProjectId,
			Endpoint:           xmaps.GetString(options.ProviderExtendedConfig, "endpoint"),
			SubAppId:           xmaps.GetInt64(options.ProviderExtendedConfig, "subAppId"),
			DomainMatchPattern: xmaps.GetString(options.ProviderExtendedConfig, "domainMatchPattern"),
			Domain:             xmaps.GetString(options.ProviderExtendedConfig, "domain"),
		})
		return provider, err
	})
}
