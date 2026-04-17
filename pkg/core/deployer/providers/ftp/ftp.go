package ftp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/certimate-go/certimate/internal/tools/ftp"
	"github.com/certimate-go/certimate/pkg/core/deployer"
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
	// 输出证书格式。
	OutputFormat string `json:"outputFormat,omitempty"`
	// 输出私钥文件路径。
	OutputKeyPath string `json:"outputKeyPath,omitempty"`
	// 输出证书文件路径。
	OutputCertPath string `json:"outputCertPath,omitempty"`
	// 输出服务器证书文件路径。
	// 选填。
	OutputServerCertPath string `json:"outputServerCertPath,omitempty"`
	// 输出中间证书文件路径。
	// 选填。
	OutputIntermediaCertPath string `json:"outputIntermediaCertPath,omitempty"`
	// PFX 导出密码。
	// 证书格式为 PFX 时必填。
	PfxPassword string `json:"pfxPassword,omitempty"`
	// JKS 别名。
	// 证书格式为 JKS 时必填。
	JksAlias string `json:"jksAlias,omitempty"`
	// JKS 密钥密码。
	// 证书格式为 JKS 时必填。
	JksKeypass string `json:"jksKeypass,omitempty"`
	// JKS 存储密码。
	// 证书格式为 JKS 时必填。
	JksStorepass string `json:"jksStorepass,omitempty"`
}

type Deployer struct {
	config *DeployerConfig
	logger *slog.Logger
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
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

	client, err := createFtpClient(*d.config)
	if err != nil {
		return nil, fmt.Errorf("ftp: failed to create FTP client: %w", err)
	}

	d.logger.Info("ftp connected")
	defer client.Quit(context.Background())

	// 上传证书和私钥文件
	switch d.config.OutputFormat {
	case OUTPUT_FORMAT_PEM:
		{
			if d.config.OutputKeyPath != "" {
				if err := client.MkdirAll(ctx, filepath.Dir(d.config.OutputKeyPath)); err != nil {
					return nil, fmt.Errorf("failed to upload private key file: %w", err)
				}
				if err := client.ChangeDir(ctx, filepath.Dir(d.config.OutputKeyPath)); err != nil {
					return nil, fmt.Errorf("failed to upload private key file: %w", err)
				}
				if err := client.StoreString(ctx, filepath.Base(d.config.OutputKeyPath), privkeyPEM); err != nil {
					return nil, fmt.Errorf("failed to upload private key file: %w", err)
				}
				d.logger.Info("ssl private key file uploaded", slog.String("path", d.config.OutputKeyPath))
			}

			if d.config.OutputCertPath != "" {
				if err := client.MkdirAll(ctx, filepath.Dir(d.config.OutputCertPath)); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				if err := client.ChangeDir(ctx, filepath.Dir(d.config.OutputCertPath)); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				if err := client.StoreString(ctx, filepath.Base(d.config.OutputCertPath), certPEM); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				d.logger.Info("ssl certificate file uploaded", slog.String("path", d.config.OutputCertPath))
			}

			if d.config.OutputServerCertPath != "" {
				if err := client.MkdirAll(ctx, filepath.Dir(d.config.OutputServerCertPath)); err != nil {
					return nil, fmt.Errorf("failed to upload server certificate file: %w", err)
				}
				if err := client.ChangeDir(ctx, filepath.Dir(d.config.OutputServerCertPath)); err != nil {
					return nil, fmt.Errorf("failed to upload server certificate file: %w", err)
				}
				if err := client.StoreString(ctx, filepath.Base(d.config.OutputServerCertPath), serverCertPEM); err != nil {
					return nil, fmt.Errorf("failed to upload server certificate file: %w", err)
				}
				d.logger.Info("ssl server certificate file uploaded", slog.String("path", d.config.OutputServerCertPath))
			}

			if d.config.OutputIntermediaCertPath != "" {
				if err := client.MkdirAll(ctx, filepath.Dir(d.config.OutputIntermediaCertPath)); err != nil {
					return nil, fmt.Errorf("failed to upload intermedia certificate file: %w", err)
				}
				if err := client.ChangeDir(ctx, filepath.Dir(d.config.OutputIntermediaCertPath)); err != nil {
					return nil, fmt.Errorf("failed to upload intermedia certificate file: %w", err)
				}
				if err := client.StoreString(ctx, filepath.Base(d.config.OutputIntermediaCertPath), intermediaCertPEM); err != nil {
					return nil, fmt.Errorf("failed to upload intermedia certificate file: %w", err)
				}
				d.logger.Info("ssl intermedia certificate file uploaded", slog.String("path", d.config.OutputIntermediaCertPath))
			}
		}

	case OUTPUT_FORMAT_PFX:
		{
			pfxData, err := xcert.TransformCertificateFromPEMToPFX(certPEM, privkeyPEM, d.config.PfxPassword)
			if err != nil {
				return nil, fmt.Errorf("failed to transform certificate to PFX: %w", err)
			}
			d.logger.Info("ssl certificate transformed to pfx")

			if d.config.OutputCertPath != "" {
				if err := client.MkdirAll(ctx, filepath.Dir(d.config.OutputCertPath)); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				if err := client.ChangeDir(ctx, filepath.Dir(d.config.OutputCertPath)); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				if err := client.StoreBytes(ctx, filepath.Base(d.config.OutputCertPath), pfxData); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				d.logger.Info("ssl certificate file uploaded", slog.String("path", d.config.OutputCertPath))
			}
		}

	case OUTPUT_FORMAT_JKS:
		{
			jksData, err := xcert.TransformCertificateFromPEMToJKS(certPEM, privkeyPEM, d.config.JksAlias, d.config.JksKeypass, d.config.JksStorepass)
			if err != nil {
				return nil, fmt.Errorf("failed to transform certificate to JKS: %w", err)
			}
			d.logger.Info("ssl certificate transformed to jks")

			if d.config.OutputCertPath != "" {
				if err := client.MkdirAll(ctx, filepath.Dir(d.config.OutputCertPath)); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				if err := client.ChangeDir(ctx, filepath.Dir(d.config.OutputCertPath)); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				if err := client.StoreBytes(ctx, filepath.Base(d.config.OutputCertPath), jksData); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				d.logger.Info("ssl certificate file uploaded", slog.String("path", d.config.OutputCertPath))
			}
		}

	default:
		return nil, fmt.Errorf("unsupported output format '%s'", d.config.OutputFormat)
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
