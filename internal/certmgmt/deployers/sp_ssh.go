package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/ssh"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeSSH, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForSSH{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		jumpServers := make([]dplyimpl.ServerConfig, len(credentials.JumpServers))
		for i, jumpServer := range credentials.JumpServers {
			jumpServers[i] = dplyimpl.ServerConfig{
				SshHost:          jumpServer.Host,
				SshPort:          jumpServer.Port,
				SshAuthMethod:    jumpServer.AuthMethod,
				SshUsername:      jumpServer.Username,
				SshPassword:      jumpServer.Password,
				SshKey:           jumpServer.Key,
				SshKeyPassphrase: jumpServer.KeyPassphrase,
			}
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			ServerConfig: dplyimpl.ServerConfig{
				SshHost:          credentials.Host,
				SshPort:          credentials.Port,
				SshAuthMethod:    credentials.AuthMethod,
				SshUsername:      credentials.Username,
				SshPassword:      credentials.Password,
				SshKey:           credentials.Key,
				SshKeyPassphrase: credentials.KeyPassphrase,
			},
			JumpServers:                  jumpServers,
			UseSCP:                       xmaps.GetBool(options.ProviderExtendedConfig, "useSCP"),
			PreCommand:                   xmaps.GetString(options.ProviderExtendedConfig, "preCommand"),
			PostCommand:                  xmaps.GetString(options.ProviderExtendedConfig, "postCommand"),
			FileFormat:                   xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "fileFormat", dplyimpl.FILE_FORMAT_PEM),
			FilePathForKey:               xmaps.GetString(options.ProviderExtendedConfig, "filePathForKey"),
			FilePathForCrt:               xmaps.GetString(options.ProviderExtendedConfig, "filePathForCrt"),
			FilePathForCrtOnlyServer:     xmaps.GetString(options.ProviderExtendedConfig, "filePathForCrtOnlyServer"),
			FilePathForCrtOnlyIntermedia: xmaps.GetString(options.ProviderExtendedConfig, "filePathForCrtOnlyIntermedia"),
			PfxPassword:                  xmaps.GetString(options.ProviderExtendedConfig, "pfxPassword"),
			PfxEncoder:                   xmaps.GetString(options.ProviderExtendedConfig, "pfxEncoder"),
			JksAlias:                     xmaps.GetString(options.ProviderExtendedConfig, "jksAlias"),
			JksKeypass:                   xmaps.GetString(options.ProviderExtendedConfig, "jksKeypass"),
			JksStorepass:                 xmaps.GetString(options.ProviderExtendedConfig, "jksStorepass"),
		})
		return provider, err
	})
}
