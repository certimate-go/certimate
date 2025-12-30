package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	netlify "github.com/certimate-go/certimate/pkg/core/deployer/providers/netlify"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeNetlify, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForNetlify{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := netlify.NewDeployer(&netlify.DeployerConfig{
			ApiToken:     credentials.ApiToken,
			ResourceType: xmaps.GetString(options.ProviderExtendedConfig, "resourceType"),
			SiteId:       xmaps.GetString(options.ProviderExtendedConfig, "siteId"),
		})
		return provider, err
	})
}
