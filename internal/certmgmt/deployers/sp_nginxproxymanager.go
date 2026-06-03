package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/nginxproxymanager"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeNginxProxyManager, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForNginxProxyManager{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			ServerUrl:                credentials.ServerUrl,
			AuthMethod:               credentials.AuthMethod,
			Username:                 credentials.Username,
			Password:                 credentials.Password,
			ApiToken:                 credentials.ApiToken,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
			DeployTarget:             xmaps.GetString(options.ProviderExtendedConfig, "deployTarget"),
			HostType:                 xmaps.GetString(options.ProviderExtendedConfig, "hostType"),
			HostMatchPattern:         xmaps.GetString(options.ProviderExtendedConfig, "hostMatchPattern"),
			HostId:                   xmaps.GetInt64(options.ProviderExtendedConfig, "hostId"),
			CertificateId:            xmaps.GetInt64(options.ProviderExtendedConfig, "certificateId"),
		})
		return provider, err
	})
}
