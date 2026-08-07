package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/ibmc"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeIBMC, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForIBMC{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate iBMC access config: %w", err)
		}
		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			Endpoint:                 credentials.Endpoint,
			Username:                 credentials.Username,
			Password:                 credentials.Password,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
			RestartAfterImport:       xmaps.GetOrDefaultBool(options.ProviderExtendedConfig, "restartAfterImport", true),
		})
		return provider, err
	})
}
