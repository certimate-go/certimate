package cdnfly

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	cdnflysdk "github.com/certimate-go/certimate/pkg/sdk3rd/cdnfly"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// Cdnfly 服务地址。
	ServerUrl string `json:"serverUrl"`
	// Cdnfly 用户端 API Key。
	ApiKey string `json:"apiKey"`
	// Cdnfly 用户端 API Secret。
	ApiSecret string `json:"apiSecret"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 部署目标。
	DeployTarget string `json:"deployTarget"`
	// 网站 ID。
	// 部署目标为 [DEPLOY_TARGET_WEBSITE] 时必填。
	SiteId string `json:"siteId,omitempty"`
	// 证书 ID。
	// 部署目标为 [DEPLOY_TARGET_CERTIFICATE] 时必填。
	CertificateId string `json:"certificateId,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *cdnflysdk.Client
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.ApiKey, config.ApiSecret, config.AllowInsecureConnections)
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*DeployResult, error) {
	// 根据部署目标决定业务流程
	switch d.config.DeployTarget {
	case DEPLOY_TARGET_WEBSITE:
		if err := d.deployToSite(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	case DEPLOY_TARGET_CERTIFICATE:
		if err := d.deployToCertificate(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported deploy target '%s'", d.config.DeployTarget)
	}

	return &DeployResult{}, nil
}

func (d *Deployer) deployToSite(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.SiteId == "" {
		return fmt.Errorf("config `siteId` is required")
	}

	// 获取单个网站详情
	// REF: https://doc.cdnfly.cn/wangzhanguanli-v1-sites.html#%E8%8E%B7%E5%8F%96%E5%8D%95%E4%B8%AA%E7%BD%91%E7%AB%99%E8%AF%A6%E6%83%85
	getSiteResp, err := d.sdkClient.GetSiteWithContext(ctx, d.config.SiteId)
	d.logger.Debug("sdk request 'GetSite'", slog.String("siteId", d.config.SiteId), slog.Any("response", getSiteResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'GetSite': %w", err)
	}

	// 添加单个证书
	// REF: https://doc.cdnfly.cn/wangzhanzhengshu-v1-certs.html#%E6%B7%BB%E5%8A%A0%E5%8D%95%E4%B8%AA%E6%88%96%E5%A4%9A%E4%B8%AA%E8%AF%81%E4%B9%A6-%E5%A4%9A%E4%B8%AA%E8%AF%81%E4%B9%A6%E6%97%B6%E6%95%B0%E6%8D%AE%E6%A0%BC%E5%BC%8F%E4%B8%BA%E6%95%B0%E7%BB%84
	createCertificateReq := &cdnflysdk.CreateCertRequest{
		Name: lo.ToPtr(fmt.Sprintf("certimate-%d", time.Now().UnixMilli())),
		Type: lo.ToPtr("custom"),
		Cert: lo.ToPtr(certPEM),
		Key:  lo.ToPtr(privkeyPEM),
	}
	createCertificateResp, err := d.sdkClient.CreateCertWithContext(ctx, createCertificateReq)
	d.logger.Debug("sdk request 'CreateCert'", slog.Any("request", createCertificateReq), slog.Any("response", createCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'CreateCert': %w", err)
	}

	// 修改单个网站
	// REF: https://doc.cdnfly.cn/wangzhanguanli-v1-sites.html#%E4%BF%AE%E6%94%B9%E5%8D%95%E4%B8%AA%E7%BD%91%E7%AB%99
	updateSiteReqHttpsListen := make(map[string]any)
	_ = json.Unmarshal([]byte(getSiteResp.Data.HttpsListen), &updateSiteReqHttpsListen)
	updateSiteReqHttpsListen["cert"] = createCertificateResp.Data
	updateSiteReq := &cdnflysdk.UpdateSiteRequest{
		HttpsListen: lo.ToPtr(updateSiteReqHttpsListen),
	}
	updateSiteResp, err := d.sdkClient.UpdateSiteWithContext(ctx, d.config.SiteId, updateSiteReq)
	d.logger.Debug("sdk request 'UpdateSite'", slog.String("siteId", d.config.SiteId), slog.Any("request", updateSiteReq), slog.Any("response", updateSiteResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'UpdateSite': %w", err)
	}

	return nil
}

func (d *Deployer) deployToCertificate(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.CertificateId == "" {
		return fmt.Errorf("config `certificateId` is required")
	}

	// 修改单个证书
	// REF: https://doc.cdnfly.cn/wangzhanzhengshu-v1-certs.html#%E4%BF%AE%E6%94%B9%E5%8D%95%E4%B8%AA%E8%AF%81%E4%B9%A6
	updateCertReq := &cdnflysdk.UpdateCertRequest{
		Type: lo.ToPtr("custom"),
		Cert: lo.ToPtr(certPEM),
		Key:  lo.ToPtr(privkeyPEM),
	}
	updateCertResp, err := d.sdkClient.UpdateCertWithContext(ctx, d.config.CertificateId, updateCertReq)
	d.logger.Debug("sdk request 'UpdateCert'", slog.String("certId", d.config.CertificateId), slog.Any("request", updateCertReq), slog.Any("response", updateCertResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'UpdateCert': %w", err)
	}

	return nil
}

func createSDKClient(serverUrl, apiKey, apiSecret string, skipTlsVerify bool) (*cdnflysdk.Client, error) {
	client, err := cdnflysdk.NewClient(serverUrl, apiKey, apiSecret)
	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, nil
}
