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
	awsacm "github.com/aws/aws-sdk-go-v2/service/acm"

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
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *awsacm.Client
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
	// REF: https://docs.aws.amazon.com/en_us/acm/latest/APIReference/API_ListCertificates.html
	// REF: https://docs.aws.amazon.com/en_us/acm/latest/APIReference/API_GetCertificate.html
	listCertificatesNextToken := (*string)(nil)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCertificatesReq := &awsacm.ListCertificatesInput{
			NextToken: listCertificatesNextToken,
			MaxItems:  aws.Int32(1000),
		}
		listCertificatesResp, err := c.sdkClient.ListCertificates(ctx, listCertificatesReq)
		c.logger.Debug("sdk request 'acm.ListCertificates'", slog.Any("request", listCertificatesReq), slog.Any("response", listCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'acm.ListCertificates': %w", err)
		}

		for _, certItem := range listCertificatesResp.CertificateSummaryList {
			// 对比证书有效期
			if certItem.NotBefore == nil || !certItem.NotBefore.Equal(certX509.NotBefore) {
				continue
			}
			if certItem.NotAfter == nil || !certItem.NotAfter.Equal(certX509.NotAfter) {
				continue
			}

			// 对比证书多域名
			if !strings.EqualFold(strings.Join(certX509.DNSNames, ","), strings.Join(certItem.SubjectAlternativeNameSummaries, ",")) {
				continue
			}

			// 对比证书内容
			getCertificateReq := &awsacm.GetCertificateInput{
				CertificateArn: certItem.CertificateArn,
			}
			getCertificateResp, err := c.sdkClient.GetCertificate(ctx, getCertificateReq)
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request 'acm.GetCertificate': %w", err)
			} else {
				if !xcert.EqualCertificatesFromPEM(certPEM, aws.ToString(getCertificateResp.Certificate)) {
					continue
				}
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			c.logger.Info("ssl certificate already exists")
			return &certmgr.UploadResult{
				CertId: *certItem.CertificateArn,
			}, nil
		}

		if len(listCertificatesResp.CertificateSummaryList) == 0 || listCertificatesResp.NextToken == nil {
			break
		}

		listCertificatesNextToken = listCertificatesResp.NextToken
	}

	// 导入证书
	// REF: https://docs.aws.amazon.com/en_us/acm/latest/APIReference/API_ImportCertificate.html
	importCertificateReq := &awsacm.ImportCertificateInput{
		Certificate:      ([]byte)(serverCertPEM),
		CertificateChain: ([]byte)(intermediaCertPEM),
		PrivateKey:       ([]byte)(privkeyPEM),
	}
	importCertificateResp, err := c.sdkClient.ImportCertificate(ctx, importCertificateReq)
	c.logger.Debug("sdk request 'acm.ImportCertificate'", slog.Any("request", importCertificateReq), slog.Any("response", importCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'acm.ImportCertificate': %w", err)
	}

	return &certmgr.UploadResult{
		CertId: aws.ToString(importCertificateResp.CertificateArn),
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	// 提取服务器证书和中间证书
	serverCertPEM, intermediaCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract certs: %w", err)
	}

	// 导入证书
	// REF: https://docs.aws.amazon.com/en_us/acm/latest/APIReference/API_ImportCertificate.html
	importCertificateReq := &awsacm.ImportCertificateInput{
		CertificateArn:   aws.String(certIdOrName),
		Certificate:      ([]byte)(serverCertPEM),
		CertificateChain: ([]byte)(intermediaCertPEM),
		PrivateKey:       ([]byte)(privkeyPEM),
	}
	importCertificateResp, err := c.sdkClient.ImportCertificate(ctx, importCertificateReq)
	c.logger.Debug("sdk request 'acm.ImportCertificate'", slog.Any("request", importCertificateReq), slog.Any("response", importCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'acm.ImportCertificate': %w", err)
	}

	return &certmgr.OperateResult{}, nil
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*awsacm.Client, error) {
	cfg, err := awscfg.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	client := awsacm.NewFromConfig(cfg, func(o *awsacm.Options) {
		o.Region = region
		o.Credentials = aws.NewCredentialsCache(awscred.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, ""))
	})
	return client, nil
}
