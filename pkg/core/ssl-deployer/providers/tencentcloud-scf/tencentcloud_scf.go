package tencentcloudscf

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcscf "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/scf/v20180416"

	"github.com/certimate-go/certimate/pkg/core"
	sslmgrsp "github.com/certimate-go/certimate/pkg/core/ssl-manager/providers/tencentcloud-ssl"
	"github.com/certimate-go/certimate/pkg/utils/ifelse"
)

type SSLDeployerProviderConfig struct {
	// 腾讯云 SecretId。
	SecretId string `json:"secretId"`
	// 腾讯云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 腾讯云接口端点。
	Endpoint string `json:"endpoint,omitempty"`
	// 腾讯云地域。
	Region string `json:"region"`
	// 自定义域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type SSLDeployerProvider struct {
	config     *SSLDeployerProviderConfig
	logger     *slog.Logger
	sdkClient  *tcscf.Client
	sslManager core.SSLManager
}

var _ core.SSLDeployer = (*SSLDeployerProvider)(nil)

func NewSSLDeployerProvider(config *SSLDeployerProviderConfig) (*SSLDeployerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl deployer provider is nil")
	}

	client, err := createSDKClient(config.SecretId, config.SecretKey, config.Endpoint, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create sdk client: %w", err)
	}

	sslmgr, err := sslmgrsp.NewSSLManagerProvider(&sslmgrsp.SSLManagerProviderConfig{
		SecretId:  config.SecretId,
		SecretKey: config.SecretKey,
		Endpoint: ifelse.
			If[string](strings.HasSuffix(strings.TrimSpace(config.Endpoint), "intl.tencentcloudapi.com")).
			Then("ssl.intl.tencentcloudapi.com"). // 国际站使用独立的接口端点
			Else(""),
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

	// 查看云函数自定义域名详情
	// REF: https://cloud.tencent.com/document/api/583/111924
	getCustomDomainReq := tcscf.NewGetCustomDomainRequest()
	getCustomDomainReq.Domain = common.StringPtr(d.config.Domain)
	getCustomDomainResp, err := d.sdkClient.GetCustomDomain(getCustomDomainReq)
	d.logger.Debug("sdk request 'scf.GetCustomDomain'", slog.Any("request", getCustomDomainReq), slog.Any("response", getCustomDomainResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'scf.GetCustomDomain': %w", err)
	}

	// 上传证书
	upres, err := d.sslManager.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 更新云函数自定义域名
	// REF: https://cloud.tencent.com/document/api/583/111922
	updateCustomDomainReq := tcscf.NewUpdateCustomDomainRequest()
	updateCustomDomainReq.Domain = common.StringPtr(d.config.Domain)
	updateCustomDomainReq.CertConfig = &tcscf.CertConf{
		CertificateId: common.StringPtr(upres.CertId),
	}
	updateCustomDomainReq.Protocol = getCustomDomainResp.Response.Protocol
	updateCustomDomainResp, err := d.sdkClient.UpdateCustomDomain(updateCustomDomainReq)
	d.logger.Debug("sdk request 'scf.UpdateCustomDomain'", slog.Any("request", updateCustomDomainReq), slog.Any("response", updateCustomDomainResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'scf.UpdateCustomDomain': %w", err)
	}

	return &core.SSLDeployResult{}, nil
}

func createSDKClient(secretId, secretKey, endpoint, region string) (*tcscf.Client, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := tcscf.NewClient(credential, region, cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
