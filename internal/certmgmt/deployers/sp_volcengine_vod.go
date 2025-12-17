package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	volcenginevod "github.com/certimate-go/certimate/pkg/core/deployer/providers/volcengine-vod"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeVolcEngineVOD, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForVolcEngine{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := volcenginevod.NewDeployer(&volcenginevod.DeployerConfig{
			AccessKeyId:        credentials.AccessKeyId,
			AccessKeySecret:    credentials.SecretAccessKey,
			SpaceName:          xmaps.GetString(options.ProviderExtendedConfig, "spaceName"),
			DomainMatchPattern: xmaps.GetString(options.ProviderExtendedConfig, "domainMatchPattern"),
			DomainType:         xmaps.GetString(options.ProviderExtendedConfig, "domainType"),
			Domain:             xmaps.GetString(options.ProviderExtendedConfig, "domain"),
		})
		return provider, err
	})
}
