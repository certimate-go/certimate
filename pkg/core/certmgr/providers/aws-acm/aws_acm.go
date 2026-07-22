package awsacm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/aws/smithy-go"

	"github.com/certimate-go/certimate/pkg/core"
	awsacmsdk "github.com/certimate-go/certimate/pkg/sdk3rd/aws/acm"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type (
	Provider      = core.Certmgr
	UploadResult  = core.CertmgrUploadResult
	ReplaceResult = core.CertmgrReplaceResult
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
	sdkClient *awsacmsdk.Client
}

var _ Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the certmgr provider is nil")
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

		listCertificatesReq := &awsacmsdk.ListCertificatesRequest{
			NextToken: listCertificatesNextToken,
			MaxItems:  aws.Int32(1000),
			SortBy:    types.SortByCreatedAt,
			SortOrder: types.SortOrderDescending,
		}
		listCertificatesResp, err := c.sdkClient.ListCertificatesWithContext(ctx, listCertificatesReq)
		c.logger.Debug("sdk request 'acm.ListCertificates'", slog.Any("request", listCertificatesReq), slog.Any("response", listCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'acm.ListCertificates': %w", err)
		}

		for _, certItem := range listCertificatesResp.CertificateSummaryList {
			// 对比证书通用名称
			// 注意，虽然文档中描述为包含了备用名称字段，但实际值不完整，因此不能用于判断证书是否相同
			if certItem.DomainName == nil || !strings.EqualFold(certX509.Subject.CommonName, *certItem.DomainName) {
				continue
			}

			// 对比证书有效期
			if certItem.NotBefore == nil || certX509.NotBefore.Unix() != certItem.NotBefore.Unix() {
				continue
			} else if certItem.NotAfter == nil || certX509.NotAfter.Unix() != certItem.NotAfter.Unix() {
				continue
			}

			// 对比证书内容
			getCertificateReq := &awsacmsdk.GetCertificateRequest{
				CertificateArn: certItem.CertificateArn,
			}
			getCertificateResp, err := c.sdkClient.GetCertificateWithContext(ctx, getCertificateReq)
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
	importCertificateReq := &awsacmsdk.ImportCertificateRequest{
		Certificate:      ([]byte)(serverCertPEM),
		CertificateChain: ([]byte)(issuerCertPEM),
		PrivateKey:       ([]byte)(privkeyPEM),
	}
	importCertificateResp, err := c.sdkClient.ImportCertificateWithContext(ctx, importCertificateReq)
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
	importCertificateReq := &awsacmsdk.ImportCertificateRequest{
		CertificateArn:   aws.String(certIdOrName),
		Certificate:      ([]byte)(serverCertPEM),
		CertificateChain: ([]byte)(issuerCertPEM),
		PrivateKey:       ([]byte)(privkeyPEM),
	}
	importCertificateResp, err := c.sdkClient.ImportCertificateWithContext(ctx, importCertificateReq)
	c.logger.Debug("sdk request 'acm.ImportCertificate'", slog.Any("request", importCertificateReq), slog.Any("response", importCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'acm.ImportCertificate': %w", err)
	}

	return &ReplaceResult{}, nil
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*awsacmsdk.Client, error) {
	client, err := awsacmsdk.NewClient(
		awsacmsdk.WithAkSk(accessKeyId, secretAccessKey),
		awsacmsdk.WithRegion(region),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
