package deployers

import (
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	baotapanelgo "github.com/certimate-go/certimate/pkg/core/deployer/providers/baotapanelgo"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeBaotaPanelGo, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForBaotaPanelGo{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := baotapanelgo.NewDeployer(&baotapanelgo.DeployerConfig{
			ServerUrl:                credentials.ServerUrl,
			ApiKey:                   credentials.ApiKey,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
			SiteType:                 xmaps.GetString(options.ProviderExtendedConfig, "siteType"),
			SiteNames:                lo.Filter(strings.Split(xmaps.GetString(options.ProviderExtendedConfig, "siteNames"), ";"), func(s string, _ int) bool { return s != "" }),
		})
		return provider, err
	})
}
