package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/huaweiibmc"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeHuaweiIBMC, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForHuaweiIBMC{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate Huawei iBMC access config: %w", err)
		}
		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			Host:                     credentials.Host,
			Username:                 credentials.Username,
			Password:                 credentials.Password,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
			AutoRestart:              xmaps.GetBool(options.ProviderExtendedConfig, "autoRestart"),
		})
		return provider, err
	})
}
