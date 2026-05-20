package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/ssh"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeSSH, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForSSH{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		jumpServers := make([]ssh.ServerConfig, len(credentials.JumpServers))
		for i, jumpServer := range credentials.JumpServers {
			jumpServers[i] = ssh.ServerConfig{
				SshHost:          jumpServer.Host,
				SshPort:          jumpServer.Port,
				SshAuthMethod:    jumpServer.AuthMethod,
				SshUsername:      jumpServer.Username,
				SshPassword:      jumpServer.Password,
				SshKey:           jumpServer.Key,
				SshKeyPassphrase: jumpServer.KeyPassphrase,
			}
		}

		provider, err := ssh.NewDeployer(&ssh.DeployerConfig{
			ServerConfig: ssh.ServerConfig{
				SshHost:          credentials.Host,
				SshPort:          credentials.Port,
				SshAuthMethod:    credentials.AuthMethod,
				SshUsername:      credentials.Username,
				SshPassword:      credentials.Password,
				SshKey:           credentials.Key,
				SshKeyPassphrase: credentials.KeyPassphrase,
			},
			JumpServers:              jumpServers,
			UseSCP:                   xmaps.GetBool(options.ProviderExtendedConfig, "useSCP"),
			PreCommand:               xmaps.GetString(options.ProviderExtendedConfig, "preCommand"),
			PostCommand:              xmaps.GetString(options.ProviderExtendedConfig, "postCommand"),
			OutputFormat:             xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "format", ssh.OUTPUT_FORMAT_PEM),
			OutputKeyPath:            xmaps.GetString(options.ProviderExtendedConfig, "keyPath"),
			OutputCertPath:           xmaps.GetString(options.ProviderExtendedConfig, "certPath"),
			OutputServerCertPath:     xmaps.GetString(options.ProviderExtendedConfig, "certPathForServerOnly"),
			OutputIntermediaCertPath: xmaps.GetString(options.ProviderExtendedConfig, "certPathForIntermediaOnly"),
			PfxPassword:              xmaps.GetString(options.ProviderExtendedConfig, "pfxPassword"),
			JksAlias:                 xmaps.GetString(options.ProviderExtendedConfig, "jksAlias"),
			JksKeypass:               xmaps.GetString(options.ProviderExtendedConfig, "jksKeypass"),
			JksStorepass:             xmaps.GetString(options.ProviderExtendedConfig, "jksStorepass"),
		})
		return provider, err
	})
}
