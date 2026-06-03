package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	tencentcloudscf "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-scf"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeTencentCloudSCF, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForTencentCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := tencentcloudscf.NewDeployer(&tencentcloudscf.DeployerConfig{
			SecretId:           credentials.SecretId,
			SecretKey:          credentials.SecretKey,
			ProjectId:          credentials.ProjectId,
			Endpoint:           xmaps.GetString(options.ProviderExtendedConfig, "endpoint"),
			Region:             xmaps.GetString(options.ProviderExtendedConfig, "region"),
			DomainMatchPattern: xmaps.GetString(options.ProviderExtendedConfig, "domainMatchPattern"),
			Domain:             xmaps.GetString(options.ProviderExtendedConfig, "domain"),
		})
		return provider, err
	})
}
