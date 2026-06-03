package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	bytepluscertcenter "github.com/certimate-go/certimate/pkg/core/deployer/providers/byteplus-certcenter"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeBytePlusCertCenter, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForBytePlus{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := bytepluscertcenter.NewDeployer(&bytepluscertcenter.DeployerConfig{
			AccessKeyId:     credentials.AccessKeyId,
			SecretAccessKey: credentials.SecretAccessKey,
			ProjectName:     credentials.ProjectName,
			Region:          xmaps.GetString(options.ProviderExtendedConfig, "region"),
		})
		return provider, err
	})
}
