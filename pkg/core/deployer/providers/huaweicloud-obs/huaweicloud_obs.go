package huaweicloudobs

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/huaweicloud-scm"
	obssdk "github.com/certimate-go/certimate/pkg/sdk3rd/huaweicloud/obs"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// 华为云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 华为云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 华为云企业项目 ID。
	EnterpriseProjectId string `json:"enterpriseProjectId,omitempty"`
	// 华为云区域。
	Region string `json:"region"`
	// 存储桶名。
	Bucket string `json:"bucket"`
	// 自定义域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *obssdk.Client
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey, config.Region, config.Bucket)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		AccessKeyId:         config.AccessKeyId,
		SecretAccessKey:     config.SecretAccessKey,
		EnterpriseProjectId: config.EnterpriseProjectId,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create certmgr: %w", err)
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
	if d.config.Region == "" {
		return nil, fmt.Errorf("config `region` is required")
	}
	if d.config.Bucket == "" {
		return nil, fmt.Errorf("config `bucket` is required")
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

	// REF: https://support.huaweicloud.com/usermanual-obs/obs_06_3200.html
	// REF: https://support.huaweicloud.com/api-obs/obs_04_0059.html
	putBucketCustomDomainReq := &obssdk.PutBucketCustomDomainRequest{
		CustomDomain:     d.config.Domain,
		Name:             upres.CertName,
		CertificateId:    upres.CertId,
		Certificate:      certPEM,
		CertificateChain: certPEM,
		PrivateKey:       privkeyPEM,
	}
	putBucketCustomDomainResp, err := d.sdkClient.PutBucketCustomDomainWithContext(ctx, putBucketCustomDomainReq)
	d.logger.Debug("sdk request 'obs.PutBucketCustomDomain'", slog.String("params.bucket", d.config.Bucket), slog.String("params.customdomain", d.config.Domain), slog.Any("request", putBucketCustomDomainReq), slog.Any("response", putBucketCustomDomainResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'obs.PutBucketCustomDomain': %w", err)
	}

	return &DeployResult{}, nil
}

func createSDKClient(accessKeyId, secretAccessKey, region, bucket string) (*obssdk.Client, error) {
	client, err := obssdk.NewClient("",
		obssdk.WithAkSk(accessKeyId, secretAccessKey),
		obssdk.WithRegion(region),
		obssdk.WithBucket(bucket),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
