package volcenginetos

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/volcengine-certcenter"
	"github.com/certimate-go/certimate/pkg/core/deployer"
)

type DeployerConfig struct {
	// 火山引擎 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 火山引擎 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 火山引擎地域。
	Region string `json:"region"`
	// 存储桶名。
	Bucket string `json:"bucket"`
	// 自定义域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *tos.ClientV2
	sdkCertmgr certmgr.Provider
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

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
		Region:          config.Region,
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM, issuerCertificatePEM string) (*deployer.DeployResult, error) {
	if d.config.Bucket == "" {
		return nil, errors.New("config `bucket` is required")
	}
	if d.config.Domain == "" {
		return nil, errors.New("config `domain` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 设置自定义域名
	// REF: https://www.volcengine.com/docs/6349/764779
	putBucketCustomDomainReq := &tos.PutBucketCustomDomainInput{
		Bucket: d.config.Bucket,
		Rule: tos.CustomDomainRule{
			Domain: d.config.Domain,
			CertID: upres.CertId,
		},
	}
	putBucketCustomDomainResp, err := d.sdkClient.PutBucketCustomDomain(ctx, putBucketCustomDomainReq)
	d.logger.Debug("sdk request 'tos.PutBucketCustomDomain'", slog.Any("request", putBucketCustomDomainReq), slog.Any("response", putBucketCustomDomainResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'tos.PutBucketCustomDomain': %w", err)
	}

	return &deployer.DeployResult{}, nil
}

func createSDKClient(accessKeyId, accessKeySecret, region string) (*tos.ClientV2, error) {
	endpoint := fmt.Sprintf("tos-%s.volces.com", region)

	client, err := tos.NewClientV2(
		endpoint,
		tos.WithRegion(region),
		tos.WithCredentials(tos.NewStaticCredentials(accessKeyId, accessKeySecret)),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
