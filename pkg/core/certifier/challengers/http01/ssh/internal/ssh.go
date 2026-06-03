package internal

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/http01"

	"github.com/certimate-go/certimate/internal/tools/ssh"
	xssh "github.com/certimate-go/certimate/pkg/utils/ssh"
)

var _ challenge.Provider = (*HTTPProvider)(nil)

type Config struct {
	ssh.Config

	UseSCP      bool
	WebRootPath string
}

func NewDefaultConfig() *Config {
	defaultCfg := ssh.NewDefaultConfig()

	return &Config{
		Config:      *defaultCfg,
		UseSCP:      false,
		WebRootPath: "/var/www/html",
	}
}

type HTTPProvider struct {
	config *Config
}

func NewHTTPProviderConfig(config *Config) (*HTTPProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	return &HTTPProvider{
		config: config,
	}, nil
}

func (p *HTTPProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	client, err := p.createSshClient()
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

func (p *HTTPProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	client, err := p.createSshClient()
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

func (p *HTTPProvider) createSshClient() (*ssh.Client, error) {
	clientCfg := ssh.NewDefaultConfig()
	clientCfg.Host = p.config.Host
	clientCfg.Port = p.config.Port
	clientCfg.AuthMethod = ssh.AuthMethodType(p.config.AuthMethod)
	clientCfg.Username = p.config.Username
	clientCfg.Password = p.config.Password
	clientCfg.Key = p.config.Key
	clientCfg.KeyPassphrase = p.config.KeyPassphrase
	for _, jumpServer := range p.config.JumpServers {
		jumpServerCfg := ssh.NewServerConfig()
		jumpServerCfg.Host = jumpServer.Host
		jumpServerCfg.Port = jumpServer.Port
		jumpServerCfg.AuthMethod = ssh.AuthMethodType(jumpServer.AuthMethod)
		jumpServerCfg.Username = jumpServer.Username
		jumpServerCfg.Password = jumpServer.Password
		jumpServerCfg.Key = jumpServer.Key
		jumpServerCfg.KeyPassphrase = jumpServer.KeyPassphrase
		clientCfg.JumpServers = append(clientCfg.JumpServers, *jumpServerCfg)
	}

	client, err := ssh.NewClient(clientCfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}
