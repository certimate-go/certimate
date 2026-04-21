package awscloudfront

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	awscred "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgracm "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aws-acm"
	mcertmgriam "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aws-iam"
	"github.com/certimate-go/certimate/pkg/core/deployer"
)

type DeployerConfig struct {
	// AWS AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// AWS SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// AWS 区域。
	Region string `json:"region"`
	// AWS CloudFront 分配 ID。
	DistributionId string `json:"distributionId"`
	// AWS CloudFront 证书来源。
	// 可取值 "ACM"、"IAM"。
	CertificateSource string `json:"certificateSource"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *cloudfront.Client
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	var pcertmgr certmgr.Provider
	switch config.CertificateSource {
	case CERTIFICATE_SOURCE_ACM:
		pcertmgr, err = mcertmgracm.NewCertmgr(&mcertmgracm.CertmgrConfig{
			AccessKeyId:     config.AccessKeyId,
			SecretAccessKey: config.SecretAccessKey,
			Region:          config.Region,
		})
		if err != nil {
			return nil, fmt.Errorf("could not create certmgr: %w", err)
		}

	case CERTIFICATE_SOURCE_IAM:
		pcertmgr, err = mcertmgriam.NewCertmgr(&mcertmgriam.CertmgrConfig{
			AccessKeyId:     config.AccessKeyId,
			SecretAccessKey: config.SecretAccessKey,
			Region:          config.Region,
			CertificatePath: "/cloudfront/",
		})
		if err != nil {
			return nil, fmt.Errorf("could not create certmgr: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported certificate source: '%s'", config.CertificateSource)
	}

	return &Deployer{
		config:     config,
		logger:     slog.Default(),
		sdkClient:  client,
		sdkCertmgr: pcertmgr,
	}, nil
}

func (d *Deployer) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}

	d.sdkCertmgr.SetLogger(logger)
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM, issuerCertificatePEM string) (*deployer.DeployResult, error) {
	if d.config.DistributionId == "" {
		return nil, errors.New("config `distribuitionId` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 获取分配配置
	// REF: https://docs.aws.amazon.com/en_us/cloudfront/latest/APIReference/API_GetDistributionConfig.html
	getDistributionConfigReq := &cloudfront.GetDistributionConfigInput{
		Id: aws.String(d.config.DistributionId),
	}
	getDistributionConfigResp, err := d.sdkClient.GetDistributionConfig(ctx, getDistributionConfigReq)
	d.logger.Debug("sdk request 'cloudfront.GetDistributionConfig'", slog.Any("request", getDistributionConfigReq), slog.Any("response", getDistributionConfigResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cloudfront.GetDistributionConfig': %w", err)
	}

	// 更新分配配置
	// REF: https://docs.aws.amazon.com/zh_cn/cloudfront/latest/APIReference/API_UpdateDistribution.html
	updateDistributionReq := &cloudfront.UpdateDistributionInput{
		Id:                 aws.String(d.config.DistributionId),
		DistributionConfig: getDistributionConfigResp.DistributionConfig,
		IfMatch:            getDistributionConfigResp.ETag,
	}
	if updateDistributionReq.DistributionConfig.ViewerCertificate == nil {
		updateDistributionReq.DistributionConfig.ViewerCertificate = &types.ViewerCertificate{}
	}
	updateDistributionReq.DistributionConfig.ViewerCertificate.CloudFrontDefaultCertificate = aws.Bool(false)
	switch d.config.CertificateSource {
	case CERTIFICATE_SOURCE_ACM:
		updateDistributionReq.DistributionConfig.ViewerCertificate.ACMCertificateArn = aws.String(upres.CertId)
		updateDistributionReq.DistributionConfig.ViewerCertificate.IAMCertificateId = nil

	case CERTIFICATE_SOURCE_IAM:
		updateDistributionReq.DistributionConfig.ViewerCertificate.ACMCertificateArn = nil
		updateDistributionReq.DistributionConfig.ViewerCertificate.IAMCertificateId = aws.String(upres.CertId)
		if updateDistributionReq.DistributionConfig.ViewerCertificate.MinimumProtocolVersion == "" {
			updateDistributionReq.DistributionConfig.ViewerCertificate.MinimumProtocolVersion = types.MinimumProtocolVersionTLSv122018
		}
		if updateDistributionReq.DistributionConfig.ViewerCertificate.SSLSupportMethod == "" {
			updateDistributionReq.DistributionConfig.ViewerCertificate.SSLSupportMethod = types.SSLSupportMethodSniOnly
		}
	}
	updateDistributionResp, err := d.sdkClient.UpdateDistribution(ctx, updateDistributionReq)
	d.logger.Debug("sdk request 'cloudfront.UpdateDistribution'", slog.Any("request", updateDistributionReq), slog.Any("response", updateDistributionResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cloudfront.UpdateDistribution': %w", err)
	}

	return &deployer.DeployResult{}, nil
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*cloudfront.Client, error) {
	cfg, err := awscfg.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	client := cloudfront.NewFromConfig(cfg, func(o *cloudfront.Options) {
		o.Region = region
		o.Credentials = aws.NewCredentialsCache(awscred.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, ""))
	})
	return client, nil
}
