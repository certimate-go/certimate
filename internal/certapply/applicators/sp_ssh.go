package applicators

import (
	"fmt"

	"github.com/go-acme/lego/v4/challenge"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-http01/providers/ssh"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	register := func(providerType domain.ACMEHttp01ProviderType) {
		if err := ACMEHttp01Registries.Register(providerType, func(options *ProviderFactoryOptions) (challenge.Provider, error) {
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

			provider, err := ssh.NewChallengeProvider(&ssh.ChallengeProviderConfig{
				ServerConfig: ssh.ServerConfig{
					SshHost:          credentials.Host,
					SshPort:          credentials.Port,
					SshAuthMethod:    credentials.AuthMethod,
					SshUsername:      credentials.Username,
					SshPassword:      credentials.Password,
					SshKey:           credentials.Key,
					SshKeyPassphrase: credentials.KeyPassphrase,
				},
				JumpServers: jumpServers,
				UseSCP:      xmaps.GetBool(options.ProviderExtendedConfig, "useSCP"),
				WebRootPath: xmaps.GetString(options.ProviderExtendedConfig, "webRootPath"),
			})
			return provider, err
		}); err != nil {
			panic(err)
		}
	}

	register(domain.ACMEHttp01ProviderTypeSSH)
	register(domain.ACMEHttp01ProviderTypeDockerHost)
}
