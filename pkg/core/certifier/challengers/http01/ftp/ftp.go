package ftp

import (
	"fmt"

	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/certifier/challengers/http01/ftp/internal"
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

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := internal.NewDefaultConfig()
	providerConfig.Host = config.FtpHost
	providerConfig.Port = int(config.FtpPort)
	providerConfig.Username = config.FtpUsername
	providerConfig.Password = config.FtpPassword
	providerConfig.WebRootPath = config.WebRootPath

	provider, err := internal.NewHTTPProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
