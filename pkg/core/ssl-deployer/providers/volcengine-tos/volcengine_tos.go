package volcenginetos

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"

	"github.com/certimate-go/certimate/pkg/core"
	sslmgrsp "github.com/certimate-go/certimate/pkg/core/ssl-manager/providers/volcengine-certcenter"
)

type SSLDeployerProviderConfig struct {
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

type SSLDeployerProvider struct {
	config     *SSLDeployerProviderConfig
	logger     *slog.Logger
	sdkClient  *tos.ClientV2
	sslManager core.SSLManager
}

var _ core.SSLDeployer = (*SSLDeployerProvider)(nil)

func NewSSLDeployerProvider(config *SSLDeployerProviderConfig) (*SSLDeployerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create sdk client: %w", err)
	}

	sslmgr, err := sslmgrsp.NewSSLManagerProvider(&sslmgrsp.SSLManagerProviderConfig{
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
		Region:          config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create ssl manager: %w", err)
	}

	return &SSLDeployerProvider{
		config:     config,
		logger:     slog.Default(),
		sdkClient:  client,
		sslManager: sslmgr,
	}, nil
}

func (d *SSLDeployerProvider) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}

	d.sslManager.SetLogger(logger)
}

func (d *SSLDeployerProvider) Deploy(ctx context.Context, certPEM string, privkeyPEM string) (*core.SSLDeployResult, error) {
	if d.config.Bucket == "" {
		return nil, errors.New("config `bucket` is required")
	}
	if d.config.Domain == "" {
		return nil, errors.New("config `domain` is required")
	}

	// 上传证书
	upres, err := d.sslManager.Upload(ctx, certPEM, privkeyPEM)
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
	putBucketCustomDomainResp, err := d.sdkClient.PutBucketCustomDomain(context.TODO(), putBucketCustomDomainReq)
	d.logger.Debug("sdk request 'tos.PutBucketCustomDomain'", slog.Any("request", putBucketCustomDomainReq), slog.Any("response", putBucketCustomDomainResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'tos.PutBucketCustomDomain': %w", err)
	}

	return &core.SSLDeployResult{}, nil
}

func createSDKClient(accessKeyId, accessKeySecret, region string) (*tos.ClientV2, error) {
	endpoint := fmt.Sprintf("tos-%s.ivolces.com", region)

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
