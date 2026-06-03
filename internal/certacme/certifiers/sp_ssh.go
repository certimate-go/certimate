package certifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	chlgimpl "github.com/certimate-go/certimate/pkg/core/certifier/challengers/http01/ssh"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	ACMEHttp01Registries.MustRegister(domain.ACMEHttp01ProviderTypeSSH, func(options *ProviderFactoryOptions) (core.ACMEChallenger, error) {
		credentials := domain.AccessConfigForSSH{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		jumpServers := make([]chlgimpl.ServerConfig, len(credentials.JumpServers))
		for i, jumpServer := range credentials.JumpServers {
			jumpServers[i] = chlgimpl.ServerConfig{
				SshHost:          jumpServer.Host,
				SshPort:          jumpServer.Port,
				SshAuthMethod:    jumpServer.AuthMethod,
				SshUsername:      jumpServer.Username,
				SshPassword:      jumpServer.Password,
				SshKey:           jumpServer.Key,
				SshKeyPassphrase: jumpServer.KeyPassphrase,
			}
		}

		provider, err := chlgimpl.NewChallenger(&chlgimpl.ChallengerConfig{
			ServerConfig: chlgimpl.ServerConfig{
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
	})
}
