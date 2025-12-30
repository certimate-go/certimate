package deployers

import (
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	baotapanel "github.com/certimate-go/certimate/pkg/core/deployer/providers/baotapanel"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeBaotaPanel, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForBaotaPanel{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := baotapanel.NewDeployer(&baotapanel.DeployerConfig{
			ServerUrl:                credentials.ServerUrl,
			ApiKey:                   credentials.ApiKey,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
			SiteType:                 xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "siteType", "other"),
			SiteNames:                lo.Filter(strings.Split(xmaps.GetString(options.ProviderExtendedConfig, "siteNames"), ";"), func(s string, _ int) bool { return s != "" }),
		})
		return provider, err
	})
}
