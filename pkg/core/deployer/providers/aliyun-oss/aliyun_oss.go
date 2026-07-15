package aliyunoss

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aliyun-cas"
	osssdk "github.com/certimate-go/certimate/pkg/sdk3rd/alibabacloud/oss"
	xalibabacloud "github.com/certimate-go/certimate/pkg/utils/third-party/alibabacloud"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
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
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *osssdk.Client
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret, config.Region, config.Bucket)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
		ResourceGroupId: config.ResourceGroupId,
		Region:          lo.Ternary(xalibabacloud.IsIntlRegion(config.Region), "ap-southeast-1", ""),
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

	// 为存储空间绑定自定义域名
	// REF: https://help.aliyun.com/zh/oss/developer-reference/putcname
	putBucketCnameReq := &osssdk.PutCnameRequest{
		Cname: &osssdk.PutCnameRequestCname{
			Domain: tea.String(d.config.Domain),
			CertificateConfiguration: &osssdk.PutCnameRequestCnameCertificateConfiguration{
				CertId:      tea.String(upres.ExtendedData["CertIdWithRegion"].(string)),
				Certificate: tea.String(certPEM),
				PrivateKey:  tea.String(privkeyPEM),
				Force:       tea.Bool(true),
			},
		},
	}
	putBucketCnameResp, err := d.sdkClient.PutBucketCnameWithContext(ctx, putBucketCnameReq)
	d.logger.Debug("sdk request 'oss.PutBucketCname'", slog.String("params.bucket", d.config.Bucket), slog.Any("request", putBucketCnameReq), slog.Any("response", putBucketCnameResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'oss.PutBucketCname': %w", err)
	}

	return &DeployResult{}, nil
}

func createSDKClient(accessKeyId, accessKeySecret, region, bucket string) (*osssdk.Client, error) {
	client, err := osssdk.NewClient("",
		osssdk.WithAkSk(accessKeyId, accessKeySecret),
		osssdk.WithRegion(region),
		osssdk.WithBucket(bucket),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
