package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	zenlayercdn "github.com/certimate-go/certimate/pkg/core/deployer/providers/zenlayer-cdn"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeZenlayerCDN, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForZenlayer{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := zenlayercdn.NewDeployer(&zenlayercdn.DeployerConfig{
			AccessKeyId:        credentials.AccessKeyId,
			AccessKeyPassword:  credentials.AccessKeyPassword,
			ResourceGroupId:    credentials.ResourceGroupId,
			ResourceType:       xmaps.GetString(options.ProviderExtendedConfig, "resourceType"),
			DomainMatchPattern: xmaps.GetString(options.ProviderExtendedConfig, "domainMatchPattern"),
			Domain:             xmaps.GetString(options.ProviderExtendedConfig, "domain"),
			CertificateId:      xmaps.GetString(options.ProviderExtendedConfig, "certificateId"),
		})
		return provider, err
	})
}
