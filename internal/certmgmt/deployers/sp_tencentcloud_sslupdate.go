package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-ssl-update"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeTencentCloudSSLUpdate, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForTencentCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			SecretId:         credentials.SecretId,
			SecretKey:        credentials.SecretKey,
			ProjectId:        credentials.ProjectId,
			Endpoint:         xmaps.GetString(options.ProviderExtendedConfig, "endpoint"),
			CertificateId:    xmaps.GetString(options.ProviderExtendedConfig, "certificateId"),
			ResourceRegions:  xmaps.GetStringsBySplit(options.ProviderExtendedConfig, "resourceRegions", ";"),
			ResourceProducts: xmaps.GetStringsBySplit(options.ProviderExtendedConfig, "resourceProducts", ";"),
			IsReplaced:       xmaps.GetBool(options.ProviderExtendedConfig, "isReplaced"),
		})
		return provider, err
	})
}
