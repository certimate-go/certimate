package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	nginxproxymanager "github.com/certimate-go/certimate/pkg/core/deployer/providers/nginxproxymanager"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeNginxProxyManager, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForNginxProxyManager{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := nginxproxymanager.NewDeployer(&nginxproxymanager.DeployerConfig{
			ServerUrl:                credentials.ServerUrl,
			AuthMethod:               credentials.AuthMethod,
			Username:                 credentials.Username,
			Password:                 credentials.Password,
			ApiToken:                 credentials.ApiToken,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
			ResourceType:             xmaps.GetString(options.ProviderExtendedConfig, "resourceType"),
			HostType:                 xmaps.GetString(options.ProviderExtendedConfig, "hostType"),
			HostMatchPattern:         xmaps.GetString(options.ProviderExtendedConfig, "hostMatchPattern"),
			HostId:                   xmaps.GetInt64(options.ProviderExtendedConfig, "hostId"),
			CertificateId:            xmaps.GetInt64(options.ProviderExtendedConfig, "certificateId"),
		})
		return provider, err
	})
}
