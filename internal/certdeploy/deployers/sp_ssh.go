package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/ssh"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	register := func(providerType domain.DeploymentProviderType) {
		if err := Registries.Register(providerType, func(options *ProviderFactoryOptions) (core.SSLDeployer, error) {
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

			provider, err := ssh.NewSSLDeployerProvider(&ssh.SSLDeployerProviderConfig{
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
				OutputFormat:             ssh.OutputFormatType(xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "format", string(ssh.OUTPUT_FORMAT_PEM))),
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
		}); err != nil {
			panic(err)
		}
	}

	register(domain.DeploymentProviderTypeSSH)
	register(domain.DeploymentProviderTypeDockerHost)
}
