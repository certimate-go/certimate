package aliyunddos

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	aliddos "github.com/alibabacloud-go/ddoscoo-20200101/v4/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/certimate-go/certimate/pkg/core"
	sslmgrsp "github.com/certimate-go/certimate/pkg/core/ssl-manager/providers/aliyun-cas"
	"github.com/certimate-go/certimate/pkg/utils/ifelse"
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
	// 网站域名（支持泛域名）。
	Domain string `json:"domain"`
}

type SSLDeployerProvider struct {
	config     *SSLDeployerProviderConfig
	logger     *slog.Logger
	sdkClient  *aliddos.Client
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
		ResourceGroupId: config.ResourceGroupId,
		Region: ifelse.
			If[string](config.Region == "" || strings.HasPrefix(config.Region, "cn-")).
			Then("cn-hangzhou").
			Else("ap-southeast-1"),
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
		return nil, errors.New("config `domain` is required")
	}

	// 上传证书
	upres, err := d.sslManager.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 为网站业务转发规则关联 SSL 证书
	// REF: https://help.aliyun.com/zh/anti-ddos/anti-ddos-pro-and-premium/developer-reference/api-ddoscoo-2020-01-01-associatewebcert
	certId, _ := strconv.Atoi(upres.CertId)
	associateWebCertReq := &aliddos.AssociateWebCertRequest{
		Domain: tea.String(d.config.Domain),
		CertId: tea.Int32(int32(certId)),
	}
	associateWebCertResp, err := d.sdkClient.AssociateWebCert(associateWebCertReq)
	d.logger.Debug("sdk request 'dcdn.AssociateWebCert'", slog.Any("request", associateWebCertReq), slog.Any("response", associateWebCertResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'dcdn.AssociateWebCert': %w", err)
	}

	return &core.SSLDeployResult{}, nil
}

func createSDKClient(accessKeyId, accessKeySecret, region string) (*aliddos.Client, error) {
	// 接入点一览 https://api.aliyun.com/product/ddoscoo
	var endpoint string
	switch region {
	case "":
		endpoint = "ddoscoo.cn-hangzhou.aliyuncs.com"
	default:
		endpoint = fmt.Sprintf("ddoscoo.%s.aliyuncs.com", region)
	}

	config := &aliopen.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(endpoint),
	}

	client, err := aliddos.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
