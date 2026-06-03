package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/baishan-cdn"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeBaishanCDN, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForBaishan{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			ApiToken:           credentials.ApiToken,
			DeployTarget:       xmaps.GetString(options.ProviderExtendedConfig, "deployTarget"),
			DomainMatchPattern: xmaps.GetString(options.ProviderExtendedConfig, "domainMatchPattern"),
			Domain:             xmaps.GetString(options.ProviderExtendedConfig, "domain"),
			CertificateId:      xmaps.GetString(options.ProviderExtendedConfig, "certificateId"),
		})
		return provider, err
	})
}
