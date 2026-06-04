package awsapigateway

import (
	"context"
	"fmt"
	"log/slog"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	awscred "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimplacm "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aws-acm"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// AWS AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// AWS SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// AWS 区域。
	Region string `json:"region"`
	// AWS API 网关自定义域名（支持泛域名）。
	Domain string `json:"domain"`
	// AWS API 网关证书来源。
	// 可取值 "ACM"。
	CertificateSource string `json:"certificateSource"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *apigatewayv2.Client
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	var pcertmgr core.Certmgr
	switch config.CertificateSource {
	case CERTIFICATE_SOURCE_ACM:
		pcertmgr, err = cmgrimplacm.NewCertmgr(&cmgrimplacm.CertmgrConfig{
			AccessKeyId:     config.AccessKeyId,
			SecretAccessKey: config.SecretAccessKey,
			Region:          config.Region,
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*DeployResult, error) {
	if d.config.Domain == "" {
		return nil, fmt.Errorf("config `domain` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 更新自定义域名
	// REF: https://docs.aws.amazon.com/apigateway/latest/api/API_UpdateDomainName.html
	updateDomainNameReq := &apigatewayv2.UpdateDomainNameInput{
		DomainName: aws.String(d.config.Domain),
		DomainNameConfigurations: []types.DomainNameConfiguration{
			{
				CertificateArn: aws.String(upres.ExtendedData["Arn"].(string)),
			},
		},
	}
	updateDomainNameResp, err := d.sdkClient.UpdateDomainName(ctx, updateDomainNameReq)
	d.logger.Debug("sdk request 'apigatewayv2.UpdateDomainName'", slog.Any("request", updateDomainNameReq), slog.Any("response", updateDomainNameResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'apigatewayv2.UpdateDomainName': %w", err)
	}

	return &DeployResult{}, nil
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*apigatewayv2.Client, error) {
	cfg, err := awscfg.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	client := apigatewayv2.NewFromConfig(cfg, func(o *apigatewayv2.Options) {
		o.Region = region
		o.Credentials = aws.NewCredentialsCache(awscred.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, ""))
	})
	return client, nil
}
