package ctcccloudfaas

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	ctyunfaas "github.com/certimate-go/certimate/pkg/sdk3rd/ctyun/faas"
)

type DeployerConfig struct {
	// 天翼云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 天翼云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 天翼云资源池 ID。
	RegionId string `json:"regionId"`
	// 自定义域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *ctyunfaas.Client
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return &Deployer{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (d *Deployer) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	if d.config.RegionId == "" {
		return nil, errors.New("config `regionId` is required")
	}
	if d.config.Domain == "" {
		return nil, errors.New("config `domain` is required")
	}

	// 获取自定义域名配置
	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=53&api=16002&data=42&isNormal=1&vid=40
	var faasCustomDomain *ctyunfaas.CustomDomainRecord
	getCustomDomainReq := &ctyunfaas.GetCustomDomainRequest{
		RegionId:   lo.ToPtr(d.config.RegionId),
		DomainName: lo.ToPtr(d.config.Domain),
		CnameCheck: lo.ToPtr(false),
	}
	getCustomDomainResp, err := d.sdkClient.GetCustomDomain(getCustomDomainReq)
	d.logger.Debug("sdk request 'faas.GetCustomDomain'", slog.Any("request", getCustomDomainReq), slog.Any("response", getCustomDomainResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'faas.GetCustomDomain': %w", err)
	} else {
		faasCustomDomain = getCustomDomainResp.ReturnObj

		// 已部署过此域名，跳过
		if faasCustomDomain.CertConfig != nil &&
			faasCustomDomain.CertConfig.Certificate == certPEM &&
			faasCustomDomain.CertConfig.PrivateKey == privkeyPEM {
			return &deployer.DeployResult{}, nil
		}
	}

	// 更新自定义域名
	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=53&api=16004&data=42&isNormal=1&vid=40
	updateCustomDomainReq := &ctyunfaas.UpdateCustomDomainRequest{
		RegionId:   lo.ToPtr(d.config.RegionId),
		DomainName: lo.ToPtr(d.config.Domain),
		Protocol:   lo.ToPtr(faasCustomDomain.Protocol),
		AuthConfig: faasCustomDomain.AuthConfig,
		CertConfig: &ctyunfaas.CustomDomainCertConfig{
			CertName:    fmt.Sprintf("certimate-%d", time.Now().UnixMilli()),
			Certificate: certPEM,
			PrivateKey:  privkeyPEM,
		},
	}
	if !strings.Contains(*updateCustomDomainReq.Protocol, "HTTPS") {
		if *updateCustomDomainReq.Protocol == "" {
			updateCustomDomainReq.Protocol = lo.ToPtr("HTTPS")
		} else {
			updateCustomDomainReq.Protocol = lo.ToPtr(*updateCustomDomainReq.Protocol + ",HTTPS")
		}
	}
	updateCustomDomainResp, err := d.sdkClient.UpdateCustomDomain(updateCustomDomainReq)
	d.logger.Debug("sdk request 'faas.UpdateCustomDomain'", slog.Any("request", updateCustomDomainReq), slog.Any("response", updateCustomDomainResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'faas.UpdateCustomDomain': %w", err)
	}

	return &deployer.DeployResult{}, nil
}

func createSDKClient(accessKeyId, secretAccessKey string) (*ctyunfaas.Client, error) {
	return ctyunfaas.NewClient(accessKeyId, secretAccessKey)
}
