package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	onepanel "github.com/certimate-go/certimate/pkg/core/deployer/providers/1panel"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderType1Panel, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigFor1Panel{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := onepanel.NewDeployer(&onepanel.DeployerConfig{
			ServerUrl:                credentials.ServerUrl,
			ApiVersion:               credentials.ApiVersion,
			ApiKey:                   credentials.ApiKey,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
			NodeName:                 xmaps.GetString(options.ProviderExtendedConfig, "nodeName"),
			ResourceType:             xmaps.GetString(options.ProviderExtendedConfig, "resourceType"),
			WebsiteMatchPattern:      xmaps.GetString(options.ProviderExtendedConfig, "websiteMatchPattern"),
			WebsiteId:                xmaps.GetInt64(options.ProviderExtendedConfig, "websiteId"),
			CertificateId:            xmaps.GetInt64(options.ProviderExtendedConfig, "certificateId"),
		})
		return provider, err
	})
}
