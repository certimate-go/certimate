package qiniukodo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/qiniu/go-sdk/v7/auth"

	"github.com/certimate-go/certimate/pkg/core"
	sslmgrsp "github.com/certimate-go/certimate/pkg/core/ssl-manager/providers/qiniu-sslcert"
	qiniusdk "github.com/certimate-go/certimate/pkg/sdk3rd/qiniu"
)

type SSLDeployerProviderConfig struct {
	// 七牛云 AccessKey。
	AccessKey string `json:"accessKey"`
	// 七牛云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 自定义域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type SSLDeployerProvider struct {
	config     *SSLDeployerProviderConfig
	logger     *slog.Logger
	sdkClient  *qiniusdk.KodoManager
	sslManager core.SSLManager
}

var _ core.SSLDeployer = (*SSLDeployerProvider)(nil)

func NewSSLDeployerProvider(config *SSLDeployerProviderConfig) (*SSLDeployerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl deployer provider is nil")
	}

	client := qiniusdk.NewKodoManager(auth.New(config.AccessKey, config.SecretKey))

	sslmgr, err := sslmgrsp.NewSSLManagerProvider(&sslmgrsp.SSLManagerProviderConfig{
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
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
	if d.config.Domain == "" {
		return nil, fmt.Errorf("config `domain` is required")
	}

	// 上传证书
	upres, err := d.sslManager.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 绑定空间域名证书
	bindBucketCertResp, err := d.sdkClient.BindBucketCert(context.TODO(), d.config.Domain, upres.CertId)
	d.logger.Debug("sdk request 'kodo.BindCert'", slog.String("request.domain", d.config.Domain), slog.String("request.certId", upres.CertId), slog.Any("response", bindBucketCertResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'kodo.BindCert': %w", err)
	}

	return &core.SSLDeployResult{}, nil
}
