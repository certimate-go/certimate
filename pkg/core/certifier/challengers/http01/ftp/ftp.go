package ftp

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/go-acme/lego/v4/challenge/http01"

	"github.com/certimate-go/certimate/internal/tools/ftp"
	"github.com/certimate-go/certimate/pkg/core/certifier"
)

type ChallengerConfig struct {
	// FTP 主机。
	FtpHost string `json:"ftpHost,omitempty"`
	// FTP 端口。
	// 零值时默认值 21。
	FtpPort int32 `json:"ftpPort,omitempty"`
	// FTP 登录用户名。
	FtpUsername string `json:"ftpUsername,omitempty"`
	// FTP 登录密码。
	FtpPassword string `json:"ftpPassword,omitempty"`
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
	ctx := context.Background()

	client, err := createFtpClient(*p.config)
	if err != nil {
		return fmt.Errorf("ftp: failed to create FTP client: %w", err)
	}

	defer client.Quit(ctx)

	challengePath := filepath.Join(p.config.WebRootPath, http01.ChallengePath(token))
	challengeDir := filepath.Dir(challengePath)
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

	return nil
}

func (p *provider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	client, err := createFtpClient(*p.config)
	if err != nil {
		return fmt.Errorf("ftp: failed to create FTP client: %w", err)
	}

	defer client.Quit(ctx)

	challengePath := filepath.Join(p.config.WebRootPath, http01.ChallengePath(token))
	challengeDir := filepath.Dir(challengePath)
	challengeFile := filepath.Base(challengePath)
	if err := client.ChangeDir(ctx, challengeDir); err != nil {
		return fmt.Errorf("ftp: failed to change to the \".well-known\" directory: %w", err)
	}
	if err := client.Delete(ctx, challengeFile); err != nil {
		return fmt.Errorf("ftp: failed to remove file after HTTP challenge: %w", err)
	}

	return nil
}

func createFtpClient(config ChallengerConfig) (*ftp.Client, error) {
	clientCfg := ftp.NewDefaultConfig()
	clientCfg.Host = config.FtpHost
	clientCfg.Port = int(config.FtpPort)
	clientCfg.Username = config.FtpUsername
	clientCfg.Password = config.FtpPassword

	client, err := ftp.NewClient(clientCfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}
