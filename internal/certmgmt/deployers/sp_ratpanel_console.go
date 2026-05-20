package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	ratpanelconsole "github.com/certimate-go/certimate/pkg/core/deployer/providers/ratpanel-console"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeRatPanelConsole, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForRatPanel{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := ratpanelconsole.NewDeployer(&ratpanelconsole.DeployerConfig{
			ServerUrl:                credentials.ServerUrl,
			AccessTokenId:            credentials.AccessTokenId,
			AccessToken:              credentials.AccessToken,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
		})
		return provider, err
	})
}
