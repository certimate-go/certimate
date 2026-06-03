package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/gcore-cdn"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeGcoreCDN, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForGcore{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			ApiToken:      credentials.ApiToken,
			ResourceId:    xmaps.GetInt64(options.ProviderExtendedConfig, "resourceId"),
			CertificateId: xmaps.GetInt64(options.ProviderExtendedConfig, "certificateId"),
		})
		return provider, err
	})
}
