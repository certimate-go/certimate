package tencentcloudwaf

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	certmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/tencentcloud-ssl"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	tcwaf "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/waf/v20180125"
)

type DeployerConfig struct {
	// 腾讯云 SecretId。
	SecretId string `json:"secretId"`
	// 腾讯云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 腾讯云接口端点。
	Endpoint string `json:"endpoint,omitempty"`
	// 腾讯云地域。
	Region string `json:"region"`
	// WAF 实例 ID。
	InstanceId string `json:"instanceId"`
	// 防护域名（不支持泛域名）。
	Domain string `json:"domain"`
	// 防护域名 ID。
	DomainId string `json:"domainId"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *tcwaf.Client
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.SecretId, config.SecretKey, config.Endpoint, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := certmgrimpl.NewCertmgr(&certmgrimpl.CertmgrConfig{
		SecretId:  config.SecretId,
		SecretKey: config.SecretKey,
		Endpoint: lo.
			If(strings.HasSuffix(config.Endpoint, "intl.tencentcloudapi.com"), "ssl.intl.tencentcloudapi.com"). // 国际站使用独立的接口端点
			Else(""),
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	if d.config.InstanceId == "" {
		return nil, errors.New("config `instanceId` is required")
	}
	if d.config.Domain == "" {
		return nil, errors.New("config `domain` is required")
	}
	if d.config.DomainId == "" {
		return nil, errors.New("config `domainId` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 查询单个 SaaS 型 WAF 域名详情
	// REF: https://cloud.tencent.com/document/api/627/82938
	describeDomainDetailsSaasReq := tcwaf.NewDescribeDomainDetailsSaasRequest()
	describeDomainDetailsSaasReq.InstanceId = common.StringPtr(d.config.InstanceId)
	describeDomainDetailsSaasReq.Domain = common.StringPtr(d.config.Domain)
	describeDomainDetailsSaasReq.DomainId = common.StringPtr(d.config.DomainId)
	describeDomainDetailsSaasResp, err := d.sdkClient.DescribeDomainDetailsSaasWithContext(ctx, describeDomainDetailsSaasReq)
	d.logger.Debug("sdk request 'waf.DescribeDomainDetailsSaas'", slog.Any("request", describeDomainDetailsSaasReq), slog.Any("response", describeDomainDetailsSaasResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'waf.DescribeDomainDetailsSaas': %w", err)
	}

	// 编辑 SaaS 型 WAF 域名
	// REF: https://cloud.tencent.com/document/api/627/94309
	modifySpartaProtectionReq := tcwaf.NewModifySpartaProtectionRequest()
	modifySpartaProtectionReq.InstanceID = common.StringPtr(d.config.InstanceId)
	modifySpartaProtectionReq.Domain = common.StringPtr(d.config.Domain)
	modifySpartaProtectionReq.DomainId = common.StringPtr(d.config.DomainId)
	modifySpartaProtectionReq.CertType = common.Int64Ptr(2)
	modifySpartaProtectionReq.SSLId = common.StringPtr(upres.CertId)
	modifySpartaProtectionResp, err := d.sdkClient.ModifySpartaProtectionWithContext(ctx, modifySpartaProtectionReq)
	d.logger.Debug("sdk request 'waf.ModifySpartaProtection'", slog.Any("request", modifySpartaProtectionReq), slog.Any("response", modifySpartaProtectionResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'waf.ModifySpartaProtection': %w", err)
	}

	return &deployer.DeployResult{}, nil
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
