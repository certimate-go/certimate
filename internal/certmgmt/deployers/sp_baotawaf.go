package deployers

import (
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	baotawaf "github.com/certimate-go/certimate/pkg/core/deployer/providers/baotawaf"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeBaotaWAF, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForBaotaWAF{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := baotawaf.NewDeployer(&baotawaf.DeployerConfig{
			ServerUrl:                credentials.ServerUrl,
			ApiKey:                   credentials.ApiKey,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
			SiteNames:                lo.Filter(strings.Split(xmaps.GetString(options.ProviderExtendedConfig, "siteNames"), ";"), func(s string, _ int) bool { return s != "" }),
			SitePort:                 xmaps.GetOrDefaultInt32(options.ProviderExtendedConfig, "sitePort", 443),
		})
		return provider, err
	})
}
