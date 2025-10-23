package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	baotapanelgosite "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/baotapanelgo-site"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	if err := Registries.Register(domain.DeploymentProviderTypeBaotaPanelGoSite, func(options *ProviderFactoryOptions) (core.SSLDeployer, error) {
		credentials := domain.AccessConfigForBaotaPanelGo{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := baotapanelgosite.NewSSLDeployerProvider(&baotapanelgosite.SSLDeployerProviderConfig{
			ServerUrl:                credentials.ServerUrl,
			ApiKey:                   credentials.ApiKey,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
			SiteName:                 xmaps.GetString(options.ProviderExtendedConfig, "siteName"),
		})
		return provider, err
	}); err != nil {
		panic(err)
	}
}
