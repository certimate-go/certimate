package ssh

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/tools/ssh"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/certifier/challengers/http01/ssh/internal"
)

type ServerConfig struct {
	// SSH 主机。
	SshHost string `json:"sshHost"`
	// SSH 端口。
	// 零值时默认值 22。
	SshPort int32 `json:"sshPort,omitempty"`
	// SSH 认证方式。
	// 可取值 "none"、"password"、"key"。
	// 零值时根据有无密码或私钥字段决定。
	SshAuthMethod string `json:"sshAuthMethod,omitempty"`
	// SSH 登录用户名。
	// 零值时默认值 "root"。
	SshUsername string `json:"sshUsername,omitempty"`
	// SSH 登录密码。
	SshPassword string `json:"sshPassword,omitempty"`
	// SSH 登录私钥。
	SshKey string `json:"sshKey,omitempty"`
	// SSH 登录私钥口令。
	SshKeyPassphrase string `json:"sshKeyPassphrase,omitempty"`
}

type ChallengerConfig struct {
	ServerConfig

	// 跳板机配置数组。
	JumpServers []ServerConfig `json:"jumpServers,omitempty"`
	// 是否回退使用 SCP。
	UseSCP bool `json:"useSCP,omitempty"`
	// 网站根目录路径。
	WebRootPath string `json:"webRootPath"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := internal.NewDefaultConfig()
	providerConfig.Host = config.SshHost
	providerConfig.Port = int(config.SshPort)
	providerConfig.AuthMethod = ssh.AuthMethodType(config.SshAuthMethod)
	providerConfig.Username = config.SshUsername
	providerConfig.Password = config.SshPassword
	providerConfig.Key = config.SshKey
	providerConfig.KeyPassphrase = config.SshKeyPassphrase
	for _, jumpServer := range config.JumpServers {
		jumpServerCfg := ssh.ServerConfig{
			Host:          jumpServer.SshHost,
			Port:          int(jumpServer.SshPort),
			AuthMethod:    ssh.AuthMethodType(jumpServer.SshAuthMethod),
			Username:      jumpServer.SshUsername,
			Password:      jumpServer.SshPassword,
			Key:           jumpServer.SshKey,
			KeyPassphrase: jumpServer.SshKeyPassphrase,
		}
		providerConfig.JumpServers = append(providerConfig.JumpServers, jumpServerCfg)
	}

	provider, err := internal.NewHTTPProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
