package s3

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/internal/tools/s3"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xcertpfx "github.com/certimate-go/certimate/pkg/utils/cert/pfx"
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
	// 证书格式。
	FileFormat string `json:"fileFormat"`
	// 私钥文件对象键。
	ObjectKeyForKey string `json:"objectKeyForKey,omitempty"`
	// 证书文件对象键。
	ObjectKeyForCrt string `json:"objectKeyForCrt,omitempty"`
	// 证书文件（仅含服务器证书）对象键。
	// 选填。
	ObjectKeyForCrtOnlyServer string `json:"objectKeyForCrtOnlyServer,omitempty"`
	// 证书文件（仅含中间证书）对象键。
	// 选填。
	ObjectKeyForCrtOnlyIntermedia string `json:"objectKeyForCrtOnlyIntermedia,omitempty"`
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

	// 连接到 S3
	s3Client, err := createS3Client(*d.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	// 写入证书和私钥文件
	switch d.config.FileFormat {
	case FILE_FORMAT_PEM:
		{
			if d.config.ObjectKeyForKey != "" {
				if err := s3Client.PutObjectString(ctx, d.config.Bucket, d.config.ObjectKeyForKey, privkeyPEM); err != nil {
					return nil, fmt.Errorf("failed to upload private key file: %w", err)
				}
				d.logger.Info("ssl private key file uploaded", slog.String("bucket", d.config.Bucket), slog.String("object", d.config.ObjectKeyForKey))
			}

			if d.config.ObjectKeyForCrt != "" {
				if err := s3Client.PutObjectString(ctx, d.config.Bucket, d.config.ObjectKeyForCrt, certPEM); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				d.logger.Info("ssl certificate file uploaded", slog.String("bucket", d.config.Bucket), slog.String("object", d.config.ObjectKeyForCrt))
			}

			if d.config.ObjectKeyForCrtOnlyServer != "" {
				if err := s3Client.PutObjectString(ctx, d.config.Bucket, d.config.ObjectKeyForCrtOnlyServer, serverCertPEM); err != nil {
					return nil, fmt.Errorf("failed to upload server certificate file: %w", err)
				}
				d.logger.Info("ssl server certificate file uploaded", slog.String("bucket", d.config.Bucket), slog.String("object", d.config.ObjectKeyForCrtOnlyServer))
			}

			if d.config.ObjectKeyForCrtOnlyIntermedia != "" {
				if err := s3Client.PutObjectString(ctx, d.config.Bucket, d.config.ObjectKeyForCrtOnlyIntermedia, intermediaCertPEM); err != nil {
					return nil, fmt.Errorf("failed to upload intermedia certificate file: %w", err)
				}
				d.logger.Info("ssl intermedia certificate file uploaded", slog.String("bucket", d.config.Bucket), slog.String("object", d.config.ObjectKeyForCrtOnlyIntermedia))
			}
		}

	case FILE_FORMAT_PFX:
		{
			if d.config.PfxPassword == "" {
				return nil, fmt.Errorf("config `pfxPassword` is required")
			}

			pfxEncoder, err := xcertpfx.ResolvePfxEncoder(d.config.PfxEncoder)
			if err != nil {
				return nil, fmt.Errorf("config `pfxEncoder` is invalid: %w", err)
			}

			pfxData, err := xcert.TransformCertificateFromPEMToPFX(certPEM, privkeyPEM, d.config.PfxPassword, pfxEncoder)
			if err != nil {
				return nil, fmt.Errorf("failed to transform certificate to PFX: %w", err)
			}
			d.logger.Info("ssl certificate transformed to pfx")

			if d.config.ObjectKeyForCrt != "" {
				if err := s3Client.PutObjectBytes(ctx, d.config.Bucket, d.config.ObjectKeyForCrt, pfxData); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				d.logger.Info("ssl certificate file uploaded", slog.String("bucket", d.config.Bucket), slog.String("object", d.config.ObjectKeyForCrt))
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

			if d.config.ObjectKeyForCrt != "" {
				if err := s3Client.PutObjectBytes(ctx, d.config.Bucket, d.config.ObjectKeyForCrt, jksData); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				d.logger.Info("ssl certificate file uploaded", slog.String("bucket", d.config.Bucket), slog.String("object", d.config.ObjectKeyForCrt))
			}
		}

	default:
		return nil, fmt.Errorf("unsupported file format '%s'", d.config.FileFormat)
	}

	return &deployer.DeployResult{}, nil
}

func createS3Client(config DeployerConfig) (*s3.Client, error) {
	clientCfg := s3.NewDefaultConfig()
	clientCfg.Endpoint = config.Endpoint
	clientCfg.AccessKey = config.AccessKey
	clientCfg.SecretKey = config.SecretKey
	clientCfg.SignatureVersion = config.SignatureVersion
	clientCfg.UsePathStyle = config.UsePathStyle
	clientCfg.Region = config.Region
	clientCfg.SkipTlsVerify = config.AllowInsecureConnections

	client, err := s3.NewClient(clientCfg)
	if err != nil {
		return nil, err
	}

	return client, err
}
