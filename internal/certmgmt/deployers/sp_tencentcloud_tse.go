package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-tse"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeTencentCloudTSE, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForTencentCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			SecretId:      credentials.SecretId,
			SecretKey:     credentials.SecretKey,
			ProjectId:     credentials.ProjectId,
			Endpoint:      xmaps.GetString(options.ProviderExtendedConfig, "endpoint"),
			Region:        xmaps.GetString(options.ProviderExtendedConfig, "region"),
			ServiceType:   xmaps.GetString(options.ProviderExtendedConfig, "serviceType"),
			GatewayId:     xmaps.GetString(options.ProviderExtendedConfig, "gatewayId"),
			Domains:       xmaps.GetStringsBySplit(options.ProviderExtendedConfig, "domains", ";"),
			CertificateId: xmaps.GetString(options.ProviderExtendedConfig, "certificateId"),
		})
		return provider, err
	})
}
