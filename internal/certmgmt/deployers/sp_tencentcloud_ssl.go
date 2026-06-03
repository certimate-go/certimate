package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	tencentcloudssl "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-ssl"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeTencentCloudSSL, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForTencentCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := tencentcloudssl.NewDeployer(&tencentcloudssl.DeployerConfig{
			SecretId:  credentials.SecretId,
			SecretKey: credentials.SecretKey,
			ProjectId: credentials.ProjectId,
			Endpoint:  xmaps.GetString(options.ProviderExtendedConfig, "endpoint"),
		})
		return provider, err
	})
}
