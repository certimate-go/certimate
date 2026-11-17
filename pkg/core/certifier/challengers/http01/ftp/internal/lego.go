package internal

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/http01"
	"github.com/go-acme/lego/v5/log"

	"github.com/certimate-go/certimate/internal/tools/ftp"
	xfilepath "github.com/certimate-go/certimate/pkg/utils/filepath"
)

var _ challenge.Provider = (*HTTPProvider)(nil)

type Config struct {
	ftp.Config

	WebRootPath string
}

func NewDefaultConfig() *Config {
	defaultCfg := ftp.NewDefaultConfig()

	return &Config{
		Config:      *defaultCfg,
		WebRootPath: "/",
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
		return nil, fmt.Errorf("ftp: webroot path must be set")
	}

	return &HTTPProvider{
		config: config,
	}, nil
}

func (p *HTTPProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	client, err := p.createFtpClient()
	if err != nil {
		return fmt.Errorf("ftp: failed to create FTP client: %w", err)
	}

	log.Info("ftp: ftp connected")
	defer func() {
		client.Quit()
		log.Info("ftp: ftp closed")
	}()

	challengePath := xfilepath.Join(p.config.WebRootPath, http01.ChallengePath(token))
	challengeDir := xfilepath.Dir(challengePath)
	challengeFile := filepath.Base(challengePath)
	if err := client.MkdirAll(ctx, challengeDir); err != nil {
		return fmt.Errorf("ftp: failed to create the \".well-known\" directory: %w", err)
	}
	if err := client.ChangeDir(ctx, challengeDir); err != nil {
		return fmt.Errorf("ftp: failed to change to the \".well-known\" directory: %w", err)
	}
	if err := client.StoreString(ctx, challengeFile, keyAuth); err != nil {
		return fmt.Errorf("ftp: failed to write file for HTTP challenge: %w", err)
	}

	log.Info("ftp: authz file uploaded", slog.String("path", challengePath))

	return nil
}

func (p *HTTPProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	client, err := p.createFtpClient()
	if err != nil {
		return fmt.Errorf("ftp: failed to create FTP client: %w", err)
	}

	log.Info("ftp: ftp connected")
	defer func() {
		client.Quit()
		log.Info("ftp: ftp closed")
	}()

	challengePath := xfilepath.Join(p.config.WebRootPath, http01.ChallengePath(token))
	challengeDir := xfilepath.Dir(challengePath)
	challengeFile := filepath.Base(challengePath)
	if err := client.ChangeDir(ctx, challengeDir); err != nil {
		return fmt.Errorf("ftp: failed to change to the \".well-known\" directory: %w", err)
	}
	if err := client.Delete(ctx, challengeFile); err != nil {
		return fmt.Errorf("ftp: failed to remove file after HTTP challenge: %w", err)
	}

	log.Info("ftp: authz file removed", slog.String("path", challengePath))

	return nil
}

func (p *HTTPProvider) createFtpClient() (*ftp.Client, error) {
	clientCfg := ftp.NewDefaultConfig()
	clientCfg.Host = p.config.Host
	clientCfg.Port = p.config.Port
	clientCfg.Username = p.config.Username
	clientCfg.Password = p.config.Password

	client, err := ftp.NewClient(clientCfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}
