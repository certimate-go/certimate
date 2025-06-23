package aliyunvod

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	alivod "github.com/alibabacloud-go/vod-20170321/v4/client"
	"github.com/certimate-go/certimate/pkg/core"
)

type SSLDeployerProviderConfig struct {
	// 阿里云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 阿里云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 阿里云资源组 ID。
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
	// 阿里云地域。
	Region string `json:"region"`
	// 点播加速域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type SSLDeployerProvider struct {
	config    *SSLDeployerProviderConfig
	logger    *slog.Logger
	sdkClient *alivod.Client
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

	return &SSLDeployerProvider{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (d *SSLDeployerProvider) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}
}

func (d *SSLDeployerProvider) Deploy(ctx context.Context, certPEM string, privkeyPEM string) (*core.SSLDeployResult, error) {
	if d.config.Domain == "" {
		return nil, errors.New("config `domain` is required")
	}

	// 设置域名证书
	// REF: https://help.aliyun.com/zh/vod/developer-reference/api-vod-2017-03-21-setvoddomainsslcertificate
	setVodDomainSSLCertificateReq := &alivod.SetVodDomainSSLCertificateRequest{
		DomainName:  tea.String(d.config.Domain),
		CertName:    tea.String(fmt.Sprintf("certimate-%d", time.Now().UnixMilli())),
		CertType:    tea.String("upload"),
		SSLProtocol: tea.String("on"),
		SSLPub:      tea.String(certPEM),
		SSLPri:      tea.String(privkeyPEM),
	}
	setVodDomainSSLCertificateResp, err := d.sdkClient.SetVodDomainSSLCertificate(setVodDomainSSLCertificateReq)
	d.logger.Debug("sdk request 'live.SetVodDomainSSLCertificate'", slog.Any("request", setVodDomainSSLCertificateReq), slog.Any("response", setVodDomainSSLCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'live.SetVodDomainSSLCertificate': %w", err)
	}

	return &core.SSLDeployResult{}, nil
}

func createSDKClient(accessKeyId, accessKeySecret, region string) (*alivod.Client, error) {
	// 接入点一览 https://api.aliyun.com/product/vod
	var endpoint string
	switch region {
	case "":
		endpoint = "vod.cn-hangzhou.aliyuncs.com"
	default:
		endpoint = fmt.Sprintf("vod.%s.aliyuncs.com", region)
	}

	config := &aliopen.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(endpoint),
	}

	client, err := alivod.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
