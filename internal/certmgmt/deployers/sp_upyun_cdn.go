package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/upyun-cdn"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeUpyunCDN, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForUpyun{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			Username:           credentials.Username,
			Password:           credentials.Password,
			DomainMatchPattern: xmaps.GetString(options.ProviderExtendedConfig, "domainMatchPattern"),
			Domain:             xmaps.GetString(options.ProviderExtendedConfig, "domain"),
		})
		return provider, err
	})
}
