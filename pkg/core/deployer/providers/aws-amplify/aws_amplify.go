package awsamplify

import (
	"context"
	"fmt"
	"log/slog"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amplify/types"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimplacm "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aws-acm"
	awsamplifysdk "github.com/certimate-go/certimate/pkg/sdk3rd/aws/amplify"
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
	// AWS Amplify 应用程序 ID。
	AppId string `json:"appId"`
	// AWS Amplify 自定义域名（不支持泛域名）。
	Domain string `json:"domain"`
	// AWS Amplify 证书来源。
	// 可取值 "ACM"。
	CertificateSource string `json:"certificateSource"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *awsamplifysdk.Client
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
	if d.config.AppId == "" {
		return nil, fmt.Errorf("config `appId` is required")
	}
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

	// 更新域名关联
	// REF: https://docs.aws.amazon.com/amplify/latest/APIReference/API_UpdateDomainAssociation.html
	updateDomainAssociationReq := &awsamplifysdk.UpdateDomainAssociationRequest{
		AppId:      aws.String(d.config.AppId),
		DomainName: aws.String(d.config.Domain),
		CertificateSettings: &types.CertificateSettings{
			Type:                 types.CertificateTypeCustom,
			CustomCertificateArn: aws.String(upres.ExtendedData["Arn"].(string)),
		},
	}
	updateDomainAssociationResp, err := d.sdkClient.UpdateDomainAssociationWithContext(ctx, updateDomainAssociationReq)
	d.logger.Debug("sdk request 'amplify.UpdateDomainAssociation'", slog.Any("request", updateDomainAssociationReq), slog.Any("response", updateDomainAssociationResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'amplify.UpdateDomainAssociation': %w", err)
	}

	return &DeployResult{}, nil
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*awsamplifysdk.Client, error) {
	client, err := awsamplifysdk.NewClient(
		awsamplifysdk.WithAkSk(accessKeyId, secretAccessKey),
		awsamplifysdk.WithRegion(region),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
