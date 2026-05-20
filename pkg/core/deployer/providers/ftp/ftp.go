package ftp

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/certimate-go/certimate/internal/tools/ftp"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	shared "github.com/certimate-go/certimate/pkg/core/deployer/providers/local/shared"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type DeployerConfig struct {
	// FTP 主机。
	FtpHost string `json:"ftpHost"`
	// FTP 端口。
	// 零值时默认值 21。
	FtpPort int32 `json:"ftpPort,omitempty"`
	// FTP 登录用户名。
	FtpUsername string `json:"ftpUsername,omitempty"`
	// FTP 登录密码。
	FtpPassword string `json:"ftpPassword,omitempty"`
	// 证书格式。
	FileFormat string `json:"fileFormat"`
	// 私钥文件路径。
	FilePathForKey string `json:"filePathForKey,omitempty"`
	// 证书文件路径。
	FilePathForCrt string `json:"filePathForCrt,omitempty"`
	// 证书文件（仅含服务器证书）路径。
	// 选填。
	FilePathForCrtOnlyServer string `json:"filePathForCrtOnlyServer,omitempty"`
	// 证书文件（仅含中间证书）路径。
	// 选填。
	FilePathForCrtOnlyIntermedia string `json:"filePathForCrtOnlyIntermedia,omitempty"`
	// PFX 导出密码。
	// 证书格式为 [FILE_FORMAT_PFX] 时必填。
	PfxPassword string `json:"pfxPassword,omitempty"`
	// PFX 编码器。
	// 证书格式为 [FILE_FORMAT_PFX] 时可选。
	PfxEncoder string `json:"pfxEncoder,omitempty"`
	// JKS 别名。
	// 证书格式为 [FILE_FORMAT_JKS] 时必填。
	JksAlias string `json:"jksAlias,omitempty"`
	// JKS 密钥密码。
	// 证书格式为 [FILE_FORMAT_JKS] 时必填。
	JksKeypass string `json:"jksKeypass,omitempty"`
	// JKS 存储密码。
	// 证书格式为 [FILE_FORMAT_JKS] 时必填。
	JksStorepass string `json:"jksStorepass,omitempty"`
}

type Deployer struct {
	config *DeployerConfig
	logger *slog.Logger
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	return &Deployer{
		config: config,
		logger: slog.Default(),
	}, nil
}

func (d *Deployer) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	// 提取服务器证书和中间证书
	serverCertPEM, intermediaCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract certs: %w", err)
	}

	// 连接到 FTP
	ftpClient, err := createFtpClient(*d.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create FTP client: %w", err)
	}
	defer ftpClient.Quit()
	d.logger.Info("ftp connected")

	// 上传证书和私钥文件
	switch d.config.FileFormat {
	case FILE_FORMAT_PEM:
		{
			if d.config.FilePathForKey != "" {
				if err := ftpClient.MkdirAll(ctx, filepath.Dir(d.config.FilePathForKey)); err != nil {
					return nil, fmt.Errorf("failed to upload private key file: %w", err)
				}
				if err := ftpClient.ChangeDir(ctx, filepath.Dir(d.config.FilePathForKey)); err != nil {
					return nil, fmt.Errorf("failed to upload private key file: %w", err)
				}
				if err := ftpClient.StoreString(ctx, filepath.Base(d.config.FilePathForKey), privkeyPEM); err != nil {
					return nil, fmt.Errorf("failed to upload private key file: %w", err)
				}
				d.logger.Info("ssl private key file uploaded", slog.String("path", d.config.FilePathForKey))
			}

			if d.config.FilePathForCrt != "" {
				if err := ftpClient.MkdirAll(ctx, filepath.Dir(d.config.FilePathForCrt)); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				if err := ftpClient.ChangeDir(ctx, filepath.Dir(d.config.FilePathForCrt)); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				if err := ftpClient.StoreString(ctx, filepath.Base(d.config.FilePathForCrt), certPEM); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				d.logger.Info("ssl certificate file uploaded", slog.String("path", d.config.FilePathForCrt))
			}

			if d.config.FilePathForCrtOnlyServer != "" {
				if err := ftpClient.MkdirAll(ctx, filepath.Dir(d.config.FilePathForCrtOnlyServer)); err != nil {
					return nil, fmt.Errorf("failed to upload server certificate file: %w", err)
				}
				if err := ftpClient.ChangeDir(ctx, filepath.Dir(d.config.FilePathForCrtOnlyServer)); err != nil {
					return nil, fmt.Errorf("failed to upload server certificate file: %w", err)
				}
				if err := ftpClient.StoreString(ctx, filepath.Base(d.config.FilePathForCrtOnlyServer), serverCertPEM); err != nil {
					return nil, fmt.Errorf("failed to upload server certificate file: %w", err)
				}
				d.logger.Info("ssl server certificate file uploaded", slog.String("path", d.config.FilePathForCrtOnlyServer))
			}

			if d.config.FilePathForCrtOnlyIntermedia != "" {
				if err := ftpClient.MkdirAll(ctx, filepath.Dir(d.config.FilePathForCrtOnlyIntermedia)); err != nil {
					return nil, fmt.Errorf("failed to upload intermedia certificate file: %w", err)
				}
				if err := ftpClient.ChangeDir(ctx, filepath.Dir(d.config.FilePathForCrtOnlyIntermedia)); err != nil {
					return nil, fmt.Errorf("failed to upload intermedia certificate file: %w", err)
				}
				if err := ftpClient.StoreString(ctx, filepath.Base(d.config.FilePathForCrtOnlyIntermedia), intermediaCertPEM); err != nil {
					return nil, fmt.Errorf("failed to upload intermedia certificate file: %w", err)
				}
				d.logger.Info("ssl intermedia certificate file uploaded", slog.String("path", d.config.FilePathForCrtOnlyIntermedia))
			}
		}

	case FILE_FORMAT_PFX:
		{
			if d.config.PfxPassword == "" {
				return nil, fmt.Errorf("config `pfxPassword` is required")
			}

			pfxEncoder, err := shared.ResolvePfxEncoder(d.config.PfxEncoder)
			if err != nil {
				return nil, fmt.Errorf("config `pfxEncoder` is invalid: %w", err)
			}

			pfxData, err := xcert.TransformCertificateFromPEMToPFX(certPEM, privkeyPEM, d.config.PfxPassword, pfxEncoder)
			if err != nil {
				return nil, fmt.Errorf("failed to transform certificate to PFX: %w", err)
			}
			d.logger.Info("ssl certificate transformed to pfx")

			if d.config.FilePathForCrt != "" {
				if err := ftpClient.MkdirAll(ctx, filepath.Dir(d.config.FilePathForCrt)); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				if err := ftpClient.ChangeDir(ctx, filepath.Dir(d.config.FilePathForCrt)); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				if err := ftpClient.StoreBytes(ctx, filepath.Base(d.config.FilePathForCrt), pfxData); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				d.logger.Info("ssl certificate file uploaded", slog.String("path", d.config.FilePathForCrt))
			}
		}

	case FILE_FORMAT_JKS:
		{
			if d.config.JksAlias == "" {
				return nil, fmt.Errorf("config `jksAlias` is required")
			}
			if d.config.JksKeypass == "" {
				return nil, fmt.Errorf("config `jksKeypass` is required")
			}
			if d.config.JksStorepass == "" {
				return nil, fmt.Errorf("config `jksStorepass` is required")
			}

			jksData, err := xcert.TransformCertificateFromPEMToJKS(certPEM, privkeyPEM, d.config.JksAlias, d.config.JksKeypass, d.config.JksStorepass)
			if err != nil {
				return nil, fmt.Errorf("failed to transform certificate to JKS: %w", err)
			}
			d.logger.Info("ssl certificate transformed to jks")

			if d.config.FilePathForCrt != "" {
				if err := ftpClient.MkdirAll(ctx, filepath.Dir(d.config.FilePathForCrt)); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				if err := ftpClient.ChangeDir(ctx, filepath.Dir(d.config.FilePathForCrt)); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				if err := ftpClient.StoreBytes(ctx, filepath.Base(d.config.FilePathForCrt), jksData); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				d.logger.Info("ssl certificate file uploaded", slog.String("path", d.config.FilePathForCrt))
			}
		}

	default:
		return nil, fmt.Errorf("unsupported file format '%s'", d.config.FileFormat)
	}

	return &deployer.DeployResult{}, nil
}

func createFtpClient(config DeployerConfig) (*ftp.Client, error) {
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
