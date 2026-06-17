package internal

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/http01"
	"github.com/go-acme/lego/v5/log"

	"github.com/certimate-go/certimate/internal/tools/ssh"
	xfilepath "github.com/certimate-go/certimate/pkg/utils/filepath"
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
		Config: *defaultCfg,
		UseSCP: false,
	}
}

type HTTPProvider struct {
	config *Config
}

func NewHTTPProviderConfig(config *Config) (*HTTPProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	if config.WebRootPath == "" {
		return nil, fmt.Errorf("ssh: webroot path must be set")
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

	log.Info("ssh: ssh connected")
	defer func() {
		client.Close()
		log.Info("ssh: ssh closed")
	}()

	challengePath := xfilepath.Join(p.config.WebRootPath, http01.ChallengePath(token))
	if err := xssh.WriteRemoteString(client.RawClient(), challengePath, keyAuth, p.config.UseSCP); err != nil {
		return fmt.Errorf("ssh: failed to write file for HTTP challenge: %w", err)
	}

	log.Info("ssh: authz file uploaded", slog.String("path", challengePath))

	return nil
}

func (p *HTTPProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	client, err := p.createSshClient()
	if err != nil {
		return fmt.Errorf("ssh: failed to create SSH client: %w", err)
	}

	log.Info("ssh: ssh connected")
	defer func() {
		client.Close()
		log.Info("ssh: ssh closed")
	}()

	// 删除质询文件
	challengePath := xfilepath.Join(p.config.WebRootPath, http01.ChallengePath(token))
	if err := xssh.RemoveRemote(client.RawClient(), challengePath, p.config.UseSCP); err != nil {
		return fmt.Errorf("ssh: failed to remove file after HTTP challenge: %w", err)
	}

	log.Info("ssh: authz file removed", slog.String("path", challengePath))

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
