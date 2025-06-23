package tencentcloudwaf

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcwaf "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/waf/v20180125"

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
	// 防护域名（不支持泛域名）。
	Domain string `json:"domain"`
	// 防护域名 ID。
	DomainId string `json:"domainId"`
	// 防护域名所属实例 ID。
	InstanceId string `json:"instanceId"`
}

type SSLDeployerProvider struct {
	config     *SSLDeployerProviderConfig
	logger     *slog.Logger
	sdkClient  *tcwaf.Client
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
	if d.config.DomainId == "" {
		return nil, errors.New("config `domainId` is required")
	}
	if d.config.InstanceId == "" {
		return nil, errors.New("config `instanceId` is required")
	}

	// 上传证书
	upres, err := d.sslManager.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 查询单个 SaaS 型 WAF 域名详情
	// REF: https://cloud.tencent.com/document/api/627/82938
	describeDomainDetailsSaasReq := tcwaf.NewDescribeDomainDetailsSaasRequest()
	describeDomainDetailsSaasReq.Domain = common.StringPtr(d.config.Domain)
	describeDomainDetailsSaasReq.DomainId = common.StringPtr(d.config.DomainId)
	describeDomainDetailsSaasReq.InstanceId = common.StringPtr(d.config.InstanceId)
	describeDomainDetailsSaasResp, err := d.sdkClient.DescribeDomainDetailsSaas(describeDomainDetailsSaasReq)
	d.logger.Debug("sdk request 'waf.DescribeDomainDetailsSaas'", slog.Any("request", describeDomainDetailsSaasReq), slog.Any("response", describeDomainDetailsSaasResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'waf.DescribeDomainDetailsSaas': %w", err)
	}

	// 编辑 SaaS 型 WAF 域名
	// REF: https://cloud.tencent.com/document/api/627/94309
	modifySpartaProtectionReq := tcwaf.NewModifySpartaProtectionRequest()
	modifySpartaProtectionReq.Domain = common.StringPtr(d.config.Domain)
	modifySpartaProtectionReq.DomainId = common.StringPtr(d.config.DomainId)
	modifySpartaProtectionReq.InstanceID = common.StringPtr(d.config.InstanceId)
	modifySpartaProtectionReq.CertType = common.Int64Ptr(2)
	modifySpartaProtectionReq.SSLId = common.StringPtr(upres.CertId)
	modifySpartaProtectionResp, err := d.sdkClient.ModifySpartaProtection(modifySpartaProtectionReq)
	d.logger.Debug("sdk request 'waf.ModifySpartaProtection'", slog.Any("request", modifySpartaProtectionReq), slog.Any("response", modifySpartaProtectionResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'waf.ModifySpartaProtection': %w", err)
	}

	return &core.SSLDeployResult{}, nil
}

func createSDKClient(secretId, secretKey, endpoint, region string) (*tcwaf.Client, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := tcwaf.NewClient(credential, region, cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
