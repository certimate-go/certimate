package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/cdnfly"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeCdnfly, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForCdnfly{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		deployer, err := cdnfly.NewDeployer(&cdnfly.DeployerConfig{
			ServerUrl:                credentials.ServerUrl,
			ApiKey:                   credentials.ApiKey,
			ApiSecret:                credentials.ApiSecret,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
			ResourceType:             xmaps.GetString(options.ProviderExtendedConfig, "resourceType"),
			SiteId:                   xmaps.GetString(options.ProviderExtendedConfig, "siteId"),
			CertificateId:            xmaps.GetString(options.ProviderExtendedConfig, "certificateId"),
		})
		return deployer, err
	})
}
