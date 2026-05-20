package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	uclouducdn "github.com/certimate-go/certimate/pkg/core/deployer/providers/ucloud-ucdn"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeUCloudUCDN, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForUCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := uclouducdn.NewDeployer(&uclouducdn.DeployerConfig{
			PrivateKey: credentials.PrivateKey,
			PublicKey:  credentials.PublicKey,
			ProjectId:  credentials.ProjectId,
			DomainId:   xmaps.GetString(options.ProviderExtendedConfig, "domainId"),
		})
		return provider, err
	})
}
