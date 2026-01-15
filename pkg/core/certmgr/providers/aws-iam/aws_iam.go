package awsiam

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	awscred "github.com/aws/aws-sdk-go-v2/credentials"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type CertmgrConfig struct {
	// AWS AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// AWS SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// AWS 区域。
	Region string `json:"region"`
	// IAM 证书路径。
	// 选填。
	CertificatePath string `json:"certificatePath,omitempty"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *awsiam.Client
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, errors.New("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return &Certmgr{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (c *Certmgr) SetLogger(logger *slog.Logger) {
	if logger == nil {
		c.logger = slog.New(slog.DiscardHandler)
	} else {
		c.logger = logger
	}
}

func (c *Certmgr) Upload(ctx context.Context, certPEM, privkeyPEM string) (*certmgr.UploadResult, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 提取服务器证书和中间证书
	serverCertPEM, intermediaCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract certs: %w", err)
	}

	// 获取证书列表，避免重复上传
	// REF: https://docs.aws.amazon.com/en_us/IAM/latest/APIReference/API_ListServerCertificates.html
	// REF: https://docs.aws.amazon.com/en_us/IAM/latest/APIReference/API_GetServerCertificate.html
	listServerCertificatesMarker := (*string)(nil)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listServerCertificatesReq := &awsiam.ListServerCertificatesInput{
			Marker:   listServerCertificatesMarker,
			MaxItems: aws.Int32(1000),
		}
		if c.config.CertificatePath != "" {
			listServerCertificatesReq.PathPrefix = aws.String(c.config.CertificatePath)
		}
		listServerCertificatesResp, err := c.sdkClient.ListServerCertificates(ctx, listServerCertificatesReq)
		c.logger.Debug("sdk request 'iam.ListServerCertificates'", slog.Any("request", listServerCertificatesReq), slog.Any("response", listServerCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'iam.ListServerCertificates': %w", err)
		}

		for _, certItem := range listServerCertificatesResp.ServerCertificateMetadataList {
			// 对比证书路径
			if c.config.CertificatePath != "" && aws.ToString(certItem.Path) != c.config.CertificatePath {
				continue
			}

			// 对比证书有效期
			if certItem.Expiration == nil || !certItem.Expiration.Equal(certX509.NotAfter) {
				continue
			}

			// 对比证书内容
			getServerCertificateReq := &awsiam.GetServerCertificateInput{
				ServerCertificateName: certItem.ServerCertificateName,
			}
			getServerCertificateResp, err := c.sdkClient.GetServerCertificate(ctx, getServerCertificateReq)
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request 'iam.GetServerCertificate': %w", err)
			} else {
				if !xcert.EqualCertificatesFromPEM(certPEM, aws.ToString(getServerCertificateResp.ServerCertificate.CertificateBody)) {
					continue
				}
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			c.logger.Info("ssl certificate already exists")
			return &certmgr.UploadResult{
				CertId:   aws.ToString(certItem.ServerCertificateId),
				CertName: aws.ToString(certItem.ServerCertificateName),
			}, nil
		}

		if len(listServerCertificatesResp.ServerCertificateMetadataList) == 0 || listServerCertificatesResp.Marker == nil {
			break
		}

		listServerCertificatesMarker = listServerCertificatesResp.Marker
	}

	// 生成新证书名（需符合 AWS IAM 命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// 导入证书
	// REF: https://docs.aws.amazon.com/en_us/IAM/latest/APIReference/API_UploadServerCertificate.html
	uploadServerCertificateReq := &awsiam.UploadServerCertificateInput{
		ServerCertificateName: aws.String(certName),
		Path:                  aws.String(c.config.CertificatePath),
		CertificateBody:       aws.String(serverCertPEM),
		CertificateChain:      aws.String(intermediaCertPEM),
		PrivateKey:            aws.String(privkeyPEM),
	}
	if c.config.CertificatePath == "" {
		uploadServerCertificateReq.Path = aws.String("/")
	}
	uploadServerCertificateResp, err := c.sdkClient.UploadServerCertificate(ctx, uploadServerCertificateReq)
	c.logger.Debug("sdk request 'iam.UploadServerCertificate'", slog.Any("request", uploadServerCertificateReq), slog.Any("response", uploadServerCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'iam.UploadServerCertificate': %w", err)
	}

	return &certmgr.UploadResult{
		CertId:   aws.ToString(uploadServerCertificateResp.ServerCertificateMetadata.ServerCertificateId),
		CertName: certName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	return nil, certmgr.ErrUnsupported
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*awsiam.Client, error) {
	cfg, err := awscfg.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	client := awsiam.NewFromConfig(cfg, func(o *awsiam.Options) {
		o.Region = region
		o.Credentials = aws.NewCredentialsCache(awscred.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, ""))
	})
	return client, nil
}
