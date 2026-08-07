package awsacm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	awscred "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/smithy-go"

	"github.com/certimate-go/certimate/pkg/core"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type (
	Provider      = core.Certmgr
	UploadResult  = core.CertmgrUploadResult
	ReplaceResult = core.CertmgrReplaceResult
)

type CertmgrConfig struct {
	// AWS API 认证方式。
	// 可取值 "accesskey"、"imds"。
	// 零值时默认值 [AUTH_METHOD_ACCESSKEY]。
	AuthMethod string `json:"authMethod,omitempty"`
	// AWS AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// AWS SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// AWS 区域。
	Region string `json:"region"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *acm.Client
}

var _ Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.AuthMethod, config.AccessKeyId, config.SecretAccessKey, config.Region)
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

func (c *Certmgr) Upload(ctx context.Context, certPEM, privkeyPEM string) (*UploadResult, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 提取服务器证书和中间证书
	serverCertPEM, issuerCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract certs: %w", err)
	}

	// 获取证书列表，避免重复上传
	// REF: https://docs.aws.amazon.com/acm/latest/APIReference/API_ListCertificates.html
	// REF: https://docs.aws.amazon.com/acm/latest/APIReference/API_GetCertificate.html
	listCertificatesNextToken := (*string)(nil)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCertificatesReq := &acm.ListCertificatesInput{
			NextToken: listCertificatesNextToken,
			MaxItems:  aws.Int32(1000),
		}
		listCertificatesResp, err := c.sdkClient.ListCertificates(ctx, listCertificatesReq)
		c.logger.Debug("sdk request 'acm.ListCertificates'", slog.Any("request", listCertificatesReq), slog.Any("response", listCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'acm.ListCertificates': %w", err)
		}

		for _, certItem := range listCertificatesResp.CertificateSummaryList {
			// 对比证书备用名称
			if !strings.EqualFold(strings.Join(certX509.DNSNames, ","), strings.Join(certItem.SubjectAlternativeNameSummaries, ",")) {
				continue
			}

			// 对比证书有效期
			if certItem.NotBefore == nil || !certX509.NotBefore.Equal(*certItem.NotBefore) {
				continue
			} else if certItem.NotAfter == nil || !certX509.NotAfter.Equal(*certItem.NotAfter) {
				continue
			}

			// 对比证书内容
			getCertificateReq := &acm.GetCertificateInput{
				CertificateArn: certItem.CertificateArn,
			}
			getCertificateResp, err := c.sdkClient.GetCertificate(ctx, getCertificateReq)
			if err != nil {
				var sdkErr smithy.APIError
				if errors.As(err, &sdkErr) {
					if sdkErrCode := sdkErr.ErrorCode(); sdkErrCode == "InvalidArnException" || sdkErrCode == "ResourceNotFoundException" {
						continue
					}
				}

				return nil, fmt.Errorf("failed to execute sdk request 'acm.GetCertificate': %w", err)
			} else {
				if !xcert.EqualCertificatesFromPEM(certPEM, aws.ToString(getCertificateResp.Certificate)) {
					continue
				}
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			c.logger.Info("ssl certificate already exists")
			return &UploadResult{
				CertId: aws.ToString(certItem.CertificateArn),
				ExtendedData: map[string]any{
					"Arn": aws.ToString(certItem.CertificateArn),
				},
			}, nil
		}

		if len(listCertificatesResp.CertificateSummaryList) == 0 || listCertificatesResp.NextToken == nil {
			break
		}

		listCertificatesNextToken = listCertificatesResp.NextToken
	}

	// 导入证书
	// REF: https://docs.aws.amazon.com/acm/latest/APIReference/API_ImportCertificate.html
	importCertificateReq := &acm.ImportCertificateInput{
		Certificate:      ([]byte)(serverCertPEM),
		CertificateChain: ([]byte)(issuerCertPEM),
		PrivateKey:       ([]byte)(privkeyPEM),
	}
	importCertificateResp, err := c.sdkClient.ImportCertificate(ctx, importCertificateReq)
	c.logger.Debug("sdk request 'acm.ImportCertificate'", slog.Any("request", importCertificateReq), slog.Any("response", importCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'acm.ImportCertificate': %w", err)
	}

	return &UploadResult{
		CertId: aws.ToString(importCertificateResp.CertificateArn),
		ExtendedData: map[string]any{
			"Arn": aws.ToString(importCertificateResp.CertificateArn),
		},
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*ReplaceResult, error) {
	// 提取服务器证书和中间证书
	serverCertPEM, issuerCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract certs: %w", err)
	}

	// 导入证书
	// REF: https://docs.aws.amazon.com/acm/latest/APIReference/API_ImportCertificate.html
	importCertificateReq := &acm.ImportCertificateInput{
		CertificateArn:   aws.String(certIdOrName),
		Certificate:      ([]byte)(serverCertPEM),
		CertificateChain: ([]byte)(issuerCertPEM),
		PrivateKey:       ([]byte)(privkeyPEM),
	}
	importCertificateResp, err := c.sdkClient.ImportCertificate(ctx, importCertificateReq)
	c.logger.Debug("sdk request 'acm.ImportCertificate'", slog.Any("request", importCertificateReq), slog.Any("response", importCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'acm.ImportCertificate': %w", err)
	}

	return &ReplaceResult{}, nil
}

func createSDKClient(authMethod, accessKeyId, secretAccessKey, region string) (*acm.Client, error) {
	opts := []func(options *awscfg.LoadOptions) error{
		awscfg.WithRegion(region),
	}

	staticCredsProvider := awscred.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, "")
	imdsCredsProvider := aws.NewCredentialsCache(ec2rolecreds.New())
	switch authMethod {
	case "":
		if accessKeyId != "" && secretAccessKey != "" {
			opts = append(opts, awscfg.WithCredentialsProvider(staticCredsProvider))
		}
	case AUTH_METHOD_ACCESSKEY:
		opts = append(opts, awscfg.WithCredentialsProvider(staticCredsProvider))
	case AUTH_METHOD_IMDS:
		opts = append(opts, awscfg.WithCredentialsProvider(imdsCredsProvider))
	default:
		return nil, fmt.Errorf("unsupported auth method '%s'", authMethod)
	}

	cfg, err := awscfg.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	client := acm.NewFromConfig(cfg)
	return client, nil
}
