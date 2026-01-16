package s3

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/internal/tools/s3"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type DeployerConfig struct {
	// S3 Endpoint。
	Endpoint string `json:"endpoint"`
	// S3 AccessKey。
	AccessKey string `json:"accessKey"`
	// S3 SecretKey。
	SecretKey string `json:"secretKey"`
	// S3 签名版本。
	// 可取值 "v2"、"v4"。
	// 零值时默认值 "v4"。
	SignatureVersion string `json:"signatureVersion,omitempty"`
	// 是否使用路径风格。
	UsePathStyle bool `json:"usePathStyle,omitempty"`
	// 存储区域。
	Region string `json:"region"`
	// 存储桶名。
	Bucket string `json:"bucket"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 输出证书格式。
	OutputFormat string `json:"outputFormat,omitempty"`
	// 输出证书文件路径。
	OutputCertObjectKey string `json:"outputCertObjectKey,omitempty"`
	// 输出服务器证书文件路径。
	// 选填。
	OutputServerCertObjectKey string `json:"outputServerCertObjectKey,omitempty"`
	// 输出中间证书文件路径。
	// 选填。
	OutputIntermediaCertObjectKey string `json:"outputIntermediaCertObjectKey,omitempty"`
	// 输出私钥文件路径。
	OutputKeyObjectKey string `json:"outputKeyObjectKey,omitempty"`
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
	config   *DeployerConfig
	logger   *slog.Logger
	s3Client *s3.Client
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := s3.NewClient(&s3.Config{
		Endpoint:         config.Endpoint,
		AccessKey:        config.AccessKey,
		SecretKey:        config.SecretKey,
		SignatureVersion: config.SignatureVersion,
		UsePathStyle:     config.UsePathStyle,
		Region:           config.Region,
		SkipTlsVerify:    config.AllowInsecureConnections,
	})
	if err != nil {
		return nil, fmt.Errorf("s3: failed to create s3 client: %w", err)
	}

	return &Deployer{
		config:   config,
		logger:   slog.Default(),
		s3Client: client,
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

	// 写入证书和私钥文件
	switch d.config.OutputFormat {
	case OUTPUT_FORMAT_PEM:
		{
			if err := d.s3Client.PutObjectString(ctx, d.config.Bucket, d.config.OutputCertObjectKey, certPEM); err != nil {
				return nil, fmt.Errorf("failed to upload certificate file: %w", err)
			}
			d.logger.Info("ssl certificate file uploaded", slog.String("bucket", d.config.Bucket), slog.String("object", d.config.OutputCertObjectKey))

			if d.config.OutputServerCertObjectKey != "" {
				if err := d.s3Client.PutObjectString(ctx, d.config.Bucket, d.config.OutputServerCertObjectKey, serverCertPEM); err != nil {
					return nil, fmt.Errorf("failed to upload server certificate file: %w", err)
				}
				d.logger.Info("ssl server certificate file uploaded", slog.String("bucket", d.config.Bucket), slog.String("object", d.config.OutputServerCertObjectKey))
			}

			if d.config.OutputIntermediaCertObjectKey != "" {
				if err := d.s3Client.PutObjectString(ctx, d.config.Bucket, d.config.OutputIntermediaCertObjectKey, intermediaCertPEM); err != nil {
					return nil, fmt.Errorf("failed to upload intermedia certificate file: %w", err)
				}
				d.logger.Info("ssl intermedia certificate file uploaded", slog.String("bucket", d.config.Bucket), slog.String("object", d.config.OutputIntermediaCertObjectKey))
			}

			if err := d.s3Client.PutObjectString(ctx, d.config.Bucket, d.config.OutputKeyObjectKey, privkeyPEM); err != nil {
				return nil, fmt.Errorf("failed to upload private key file: %w", err)
			}
			d.logger.Info("ssl private key file uploaded", slog.String("bucket", d.config.Bucket), slog.String("object", d.config.OutputKeyObjectKey))
		}

	case OUTPUT_FORMAT_PFX:
		{
			pfxData, err := xcert.TransformCertificateFromPEMToPFX(certPEM, privkeyPEM, d.config.PfxPassword)
			if err != nil {
				return nil, fmt.Errorf("failed to transform certificate to PFX: %w", err)
			}
			d.logger.Info("ssl certificate transformed to pfx")

			if err := d.s3Client.PutObjectBytes(ctx, d.config.Bucket, d.config.OutputCertObjectKey, pfxData); err != nil {
				return nil, fmt.Errorf("failed to upload certificate file: %w", err)
			}
			d.logger.Info("ssl certificate file uploaded", slog.String("bucket", d.config.Bucket), slog.String("object", d.config.OutputCertObjectKey))
		}

	case OUTPUT_FORMAT_JKS:
		{
			jksData, err := xcert.TransformCertificateFromPEMToJKS(certPEM, privkeyPEM, d.config.JksAlias, d.config.JksKeypass, d.config.JksStorepass)
			if err != nil {
				return nil, fmt.Errorf("failed to transform certificate to JKS: %w", err)
			}
			d.logger.Info("ssl certificate transformed to jks")

			if err := d.s3Client.PutObjectBytes(ctx, d.config.Bucket, d.config.OutputCertObjectKey, jksData); err != nil {
				return nil, fmt.Errorf("failed to upload certificate file: %w", err)
			}
			d.logger.Info("ssl certificate file uploaded", slog.String("bucket", d.config.Bucket), slog.String("object", d.config.OutputCertObjectKey))
		}

	default:
		return nil, fmt.Errorf("unsupported output format '%s'", d.config.OutputFormat)
	}

	return &deployer.DeployResult{}, nil
}
