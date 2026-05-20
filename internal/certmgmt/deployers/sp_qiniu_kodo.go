package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	qiniukodo "github.com/certimate-go/certimate/pkg/core/deployer/providers/qiniu-kodo"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeQiniuKodo, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForQiniu{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := qiniukodo.NewDeployer(&qiniukodo.DeployerConfig{
			AccessKey: credentials.AccessKey,
			SecretKey: credentials.SecretKey,
			Bucket:    xmaps.GetString(options.ProviderExtendedConfig, "bucket"),
			Domain:    xmaps.GetString(options.ProviderExtendedConfig, "domain"),
		})
		return provider, err
	})
}
