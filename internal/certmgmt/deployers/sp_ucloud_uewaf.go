package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	uclouduewaf "github.com/certimate-go/certimate/pkg/core/deployer/providers/ucloud-uewaf"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeUCloudUEWAF, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForUCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := uclouduewaf.NewDeployer(&uclouduewaf.DeployerConfig{
			PrivateKey: credentials.PrivateKey,
			PublicKey:  credentials.PublicKey,
			ProjectId:  credentials.ProjectId,
			Domain:     xmaps.GetString(options.ProviderExtendedConfig, "domain"),
		})
		return provider, err
	})
}
