package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	upyunfile "github.com/certimate-go/certimate/pkg/core/deployer/providers/upyun-file"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeUpyunFile, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForUpyun{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := upyunfile.NewDeployer(&upyunfile.DeployerConfig{
			Username: credentials.Username,
			Password: credentials.Password,
			Bucket:   xmaps.GetString(options.ProviderExtendedConfig, "bucket"),
			Domain:   xmaps.GetString(options.ProviderExtendedConfig, "domain"),
		})
		return provider, err
	})
}
