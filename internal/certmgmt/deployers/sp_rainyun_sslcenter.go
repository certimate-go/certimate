package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/rainyun-sslcenter"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeRainYunSSLCenter, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForRainYun{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			ApiKey:        credentials.ApiKey,
			CertificateId: xmaps.GetInt64(options.ProviderExtendedConfig, "certificateId"),
		})
		return provider, err
	})
}
