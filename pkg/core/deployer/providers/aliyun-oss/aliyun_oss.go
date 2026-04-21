package aliyunoss

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"

	"github.com/certimate-go/certimate/pkg/core/deployer"
)

type DeployerConfig struct {
	// 阿里云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 阿里云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 阿里云资源组 ID。
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
	// 阿里云地域。
	Region string `json:"region"`
	// 存储桶名。
	Bucket string `json:"bucket"`
	// 自定义域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *oss.Client
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return &Deployer{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (d *Deployer) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM, issuerCertificatePEM string) (*deployer.DeployResult, error) {
	if d.config.Bucket == "" {
		return nil, errors.New("config `bucket` is required")
	}
	if d.config.Domain == "" {
		return nil, errors.New("config `domain` is required")
	}

	// 为存储空间绑定自定义域名
	// REF: https://help.aliyun.com/zh/oss/developer-reference/putcname
	putCnameReq := &oss.PutCnameRequest{
		Bucket: tea.String(d.config.Bucket),
		BucketCnameConfiguration: &oss.BucketCnameConfiguration{
			Domain: tea.String(d.config.Domain),
			CertificateConfiguration: &oss.CertificateConfiguration{
				Certificate: tea.String(certPEM),
				PrivateKey:  tea.String(privkeyPEM),
				Force:       tea.Bool(true),
			},
		},
	}
	putCnameResp, err := d.sdkClient.PutCname(ctx, putCnameReq)
	d.logger.Debug("sdk request 'oss.PutCname'", slog.Any("request", putCnameReq), slog.Any("response", putCnameResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'oss.PutCname': %w", err)
	}

	return &deployer.DeployResult{}, nil
}

func createSDKClient(accessKeyId, accessKeySecret, region string) (*oss.Client, error) {
	// 接入点一览 https://api.aliyun.com/product/Oss
	var endpoint string
	switch region {
	case "":
		endpoint = "oss.aliyuncs.com"
	case
		"cn-hzjbp",
		"cn-hzjbp-a",
		"cn-hzjbp-b":
		endpoint = "oss-cn-hzjbp-a-internal.aliyuncs.com"
	case
		"cn-shanghai-finance-1",
		"cn-shenzhen-finance-1",
		"cn-beijing-finance-1",
		"cn-north-2-gov-1":
		endpoint = fmt.Sprintf("oss-%s-internal.aliyuncs.com", region)
	default:
		endpoint = fmt.Sprintf("oss-%s.aliyuncs.com", region)
	}

	provider := credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret)
	config := oss.LoadDefaultConfig().
		WithCredentialsProvider(provider).
		WithEndpoint(endpoint)
	if region != "" {
		config = config.WithRegion(region)
	}

	client := oss.NewClient(config)
	return client, nil
}
