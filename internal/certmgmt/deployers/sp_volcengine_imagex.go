package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	volcengineimagex "github.com/certimate-go/certimate/pkg/core/deployer/providers/volcengine-imagex"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeVolcEngineImageX, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForVolcEngine{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := volcengineimagex.NewDeployer(&volcengineimagex.DeployerConfig{
			AccessKeyId:     credentials.AccessKeyId,
			AccessKeySecret: credentials.SecretAccessKey,
			ProjectName:     credentials.ProjectName,
			Region:          xmaps.GetString(options.ProviderExtendedConfig, "region"),
			ServiceId:       xmaps.GetString(options.ProviderExtendedConfig, "serviceId"),
			Domain:          xmaps.GetString(options.ProviderExtendedConfig, "domain"),
		})
		return provider, err
	})
}
