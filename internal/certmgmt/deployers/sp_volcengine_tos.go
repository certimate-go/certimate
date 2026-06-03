package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	volcenginetos "github.com/certimate-go/certimate/pkg/core/deployer/providers/volcengine-tos"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeVolcEngineTOS, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForVolcEngine{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := volcenginetos.NewDeployer(&volcenginetos.DeployerConfig{
			AccessKeyId:     credentials.AccessKeyId,
			AccessKeySecret: credentials.SecretAccessKey,
			ProjectName:     credentials.ProjectName,
			Region:          xmaps.GetString(options.ProviderExtendedConfig, "region"),
			Bucket:          xmaps.GetString(options.ProviderExtendedConfig, "bucket"),
			Domain:          xmaps.GetString(options.ProviderExtendedConfig, "domain"),
		})
		return provider, err
	})
}
