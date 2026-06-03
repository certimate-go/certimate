package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/lecdn"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeLeCDN, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForLeCDN{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			ServerUrl:                credentials.ServerUrl,
			ApiVersion:               credentials.ApiVersion,
			ApiRole:                  credentials.ApiRole,
			Username:                 credentials.Username,
			Password:                 credentials.Password,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
			DeployTarget:             xmaps.GetString(options.ProviderExtendedConfig, "deployTarget"),
			CertificateId:            xmaps.GetInt64(options.ProviderExtendedConfig, "certificateId"),
			ClientId:                 xmaps.GetInt64(options.ProviderExtendedConfig, "clientId"),
		})
		return provider, err
	})
}
