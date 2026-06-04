package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/cloudflare-ssl"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeCloudflareSSL, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForCloudflare{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			ApiToken:      credentials.ApiToken,
			Environment:   xmaps.GetString(options.ProviderExtendedConfig, "environment"),
			ZoneId:        xmaps.GetString(options.ProviderExtendedConfig, "zoneId"),
			CertificateId: xmaps.GetString(options.ProviderExtendedConfig, "certificateId"),
		})
		return provider, err
	})
}
