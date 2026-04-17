package ssh

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/go-acme/lego/v4/challenge/http01"

	"github.com/certimate-go/certimate/internal/tools/ssh"
	"github.com/certimate-go/certimate/pkg/core/certifier"
	xssh "github.com/certimate-go/certimate/pkg/utils/ssh"
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

func NewChallenger(config *ChallengerConfig) (certifier.ACMEChallenger, error) {
	if config == nil {
		return nil, errors.New("the configuration of the acme challenge provider is nil")
	}

	provider := &provider{config: config}
	return provider, nil
}

type provider struct {
	config *ChallengerConfig
}

func (p *provider) Present(domain, token, keyAuth string) error {
	client, err := createSshClient(*p.config)
	if err != nil {
		return fmt.Errorf("ssh: failed to create SSH client: %w", err)
	}

	defer client.Close()

	challengePath := filepath.Join(p.config.WebRootPath, http01.ChallengePath(token))
	if err := xssh.WriteRemoteString(client.RawClient(), challengePath, keyAuth, p.config.UseSCP); err != nil {
		return fmt.Errorf("ssh: failed to write file for HTTP challenge: %w", err)
	}

	return nil
}

func (p *provider) CleanUp(domain, token, keyAuth string) error {
	client, err := createSshClient(*p.config)
	if err != nil {
		return fmt.Errorf("ssh: failed to create SSH client: %w", err)
	}

	defer client.Close()

	// 删除质询文件
	challengePath := filepath.Join(p.config.WebRootPath, http01.ChallengePath(token))
	if err := xssh.RemoveRemote(client.RawClient(), challengePath, p.config.UseSCP); err != nil {
		return fmt.Errorf("ssh: failed to remove file after HTTP challenge: %w", err)
	}

	return nil
}

func createSshClient(config ChallengerConfig) (*ssh.Client, error) {
	clientCfg := ssh.NewDefaultConfig()
	clientCfg.Host = config.SshHost
	clientCfg.Port = int(config.SshPort)
	clientCfg.AuthMethod = ssh.AuthMethodType(config.SshAuthMethod)
	clientCfg.Username = config.SshUsername
	clientCfg.Password = config.SshPassword
	clientCfg.Key = config.SshKey
	clientCfg.KeyPassphrase = config.SshKeyPassphrase
	for _, jumpServer := range config.JumpServers {
		jumpServerCfg := ssh.NewServerConfig()
		jumpServerCfg.Host = jumpServer.SshHost
		jumpServerCfg.Port = int(jumpServer.SshPort)
		jumpServerCfg.AuthMethod = ssh.AuthMethodType(jumpServer.SshAuthMethod)
		jumpServerCfg.Username = jumpServer.SshUsername
		jumpServerCfg.Password = jumpServer.SshPassword
		jumpServerCfg.Key = jumpServer.SshKey
		jumpServerCfg.KeyPassphrase = jumpServer.SshKeyPassphrase
		clientCfg.JumpServers = append(clientCfg.JumpServers, *jumpServerCfg)
	}

	client, err := ssh.NewClient(clientCfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}
