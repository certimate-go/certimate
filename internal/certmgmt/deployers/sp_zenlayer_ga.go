package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	zenlayerga "github.com/certimate-go/certimate/pkg/core/deployer/providers/zenlayer-ga"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeZenlayerGA, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForZenlayer{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := zenlayerga.NewDeployer(&zenlayerga.DeployerConfig{
			AccessKeyId:       credentials.AccessKeyId,
			AccessKeyPassword: credentials.AccessKeyPassword,
			ResourceGroupId:   credentials.ResourceGroupId,
			ResourceType:      xmaps.GetString(options.ProviderExtendedConfig, "resourceType"),
			AcceleratorId:     xmaps.GetString(options.ProviderExtendedConfig, "acceleratorId"),
			CertificateId:     xmaps.GetString(options.ProviderExtendedConfig, "certificateId"),
		})
		return provider, err
	})
}
