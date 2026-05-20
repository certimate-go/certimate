package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/proxmoxve"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeProxmoxVE, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForProxmoxVE{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := proxmoxve.NewDeployer(&proxmoxve.DeployerConfig{
			ServerUrl:                credentials.ServerUrl,
			ApiToken:                 credentials.ApiToken,
			ApiTokenSecret:           credentials.ApiTokenSecret,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
			NodeName:                 xmaps.GetString(options.ProviderExtendedConfig, "nodeName"),
			AutoRestart:              xmaps.GetBool(options.ProviderExtendedConfig, "autoRestart"),
		})
		return provider, err
	})
}
