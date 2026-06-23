package baotapanel

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	btpanelsdk "github.com/certimate-go/certimate/pkg/sdk3rd/btpanel"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// 宝塔面板服务地址。
	ServerUrl string `json:"serverUrl"`
	// 宝塔面板接口密钥。
	ApiKey string `json:"apiKey"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 网站类型。
	SiteType string `json:"siteType"`
	// 网站名称。
	SiteNames []string `json:"siteNames,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *btpanelsdk.Client
}

var _ Provider = (*Deployer)(nil)

var btProjectTypes = []string{"php", "java", "nodejs", "go", "python", "proxy", "html", "general"}

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.ApiKey, config.AllowInsecureConnections)
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
	if len(d.config.SiteNames) == 0 {
		return nil, fmt.Errorf("config `siteNames` is required")
	}

	switch d.config.SiteType {
	case "any":
		{
			// 批量更新站点证书
			// 注意，如果 v1 接口不可用，则尝试使用 v2 接口重试
			if err1 := d.updateSitesCertificateByAny(ctx, d.config.SiteNames, certPEM, privkeyPEM); err1 != nil {
				if err2 := d.updateSitesCertificateByAnyV2(ctx, d.config.SiteNames, certPEM, privkeyPEM); err2 != nil {
					return nil, errors.Join(err1, err2)
				}
			}
		}

	default:
		{
			if d.config.SiteType != "" {
				if !lo.Contains(btProjectTypes, d.config.SiteType) {
					return nil, fmt.Errorf("unsupported site type: '%s'", d.config.SiteType)
				}
			}

			// 遍历更新站点证书
			var errs []error
			for i, siteName := range d.config.SiteNames {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
					if err := d.updateSiteCertificate(ctx, d.config.SiteType, siteName, certPEM, privkeyPEM); err != nil {
						errs = append(errs, err)
					} else if i < len(d.config.SiteNames)-1 {
						xwait.DelayWithContext(ctx, 5*time.Second)
					}
				}
			}
			if len(errs) > 0 {
				return nil, errors.Join(errs...)
			}
		}
	}

	return &DeployResult{}, nil
}

func (d *Deployer) updateSiteCertificate(ctx context.Context, siteType, siteName string, certPEM, privkeyPEM string) error {
	switch siteType {
	case "proxy":
		{
			// 设置代理 SSL 证书
			modProxyComSetSSLReq := &btpanelsdk.ModProxyComSetSSLRequest{
				SiteName:    siteName,
				Certificate: certPEM,
				PrivateKey:  privkeyPEM,
			}
			modProxyComSetSSLResp, err := d.sdkClient.ModProxyComSetSSLWithContext(ctx, modProxyComSetSSLReq)
			d.logger.Debug("sdk request 'mod.proxy.com.SetSSL'", slog.Any("request", modProxyComSetSSLReq), slog.Any("response", modProxyComSetSSLResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'mod.proxy.com.SetSSL': %w", err)
			}
		}

	default:
		{
			// 设置站点 SSL 证书
			siteSetSSLReq := &btpanelsdk.SiteSetSSLRequest{
				Type:        "0",
				SiteName:    siteName,
				Certificate: certPEM,
				PrivateKey:  privkeyPEM,
			}
			siteSetSSLResp, err := d.sdkClient.SiteSetSSLWithContext(ctx, siteSetSSLReq)
			d.logger.Debug("sdk request 'site.SetSSL'", slog.Any("request", siteSetSSLReq), slog.Any("response", siteSetSSLResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'site.SetSSL': %w", err)
			}
		}
	}

	return nil
}

func (d *Deployer) updateSitesCertificateByAny(ctx context.Context, siteNames []string, certPEM, privkeyPEM string) error {
	// 上传证书
	sslCertSaveCertReq := &btpanelsdk.SSLCertSaveCertRequest{
		Certificate: certPEM,
		PrivateKey:  privkeyPEM,
	}
	sslCertSaveCertResp, err := d.sdkClient.SSLCertSaveCertWithContext(ctx, sslCertSaveCertReq)
	d.logger.Debug("sdk request 'ssl.cert.SaveCert'", slog.Any("request", sslCertSaveCertReq), slog.Any("response", sslCertSaveCertResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'ssl.cert.SaveCert': %w", err)
	}

	// 设置站点证书
	sslSetBatchCertToSiteReq := &btpanelsdk.SSLSetBatchCertToSiteRequest{
		BatchInfo: lo.Map(siteNames, func(siteName string, _ int) *btpanelsdk.SSLSetBatchCertToSiteRequestBatchInfo {
			return &btpanelsdk.SSLSetBatchCertToSiteRequestBatchInfo{
				SiteName: siteName,
				SSLHash:  sslCertSaveCertResp.SSLHash,
			}
		}),
	}
	sslSetBatchCertToSiteResp, err := d.sdkClient.SSLSetBatchCertToSiteWithContext(ctx, sslSetBatchCertToSiteReq)
	d.logger.Debug("sdk request 'ssl.SetBatchCertToSite'", slog.Any("request", sslSetBatchCertToSiteReq), slog.Any("response", sslSetBatchCertToSiteResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'ssl.SetBatchCertToSite': %w", err)
	}

	return nil
}

func (d *Deployer) updateSitesCertificateByAnyV2(ctx context.Context, siteNames []string, certPEM, privkeyPEM string) error {
	// 上传证书
	sslDomainUploadCertV2Req := &btpanelsdk.SSLDomainUploadCertV2Request{
		Certificate: certPEM,
		PrivateKey:  privkeyPEM,
	}
	sslDomainUploadCertV2Resp, err := d.sdkClient.SSLDomainUploadCertV2WithContext(ctx, sslDomainUploadCertV2Req)
	d.logger.Debug("sdk request 'v2.ssldomain.UploadCert'", slog.Any("request", sslDomainUploadCertV2Req), slog.Any("response", sslDomainUploadCertV2Resp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'v2.ssldomain.UploadCert': %w", err)
	}

	// 设置站点证书
	sslDomainCertDeploySitesV2Req := &btpanelsdk.SSLDomainCertDeploySitesV2Request{
		SSLHash: sslDomainUploadCertV2Resp.Message.SSLHash,
		Domains: siteNames,
		Append:  1,
	}
	sslDomainCertDeploySitesV2Resp, err := d.sdkClient.SSLDomainCertDeploySitesV2WithContext(ctx, sslDomainCertDeploySitesV2Req)
	d.logger.Debug("sdk request 'v2.ssldomain.CertDeploySites'", slog.Any("request", sslDomainCertDeploySitesV2Req), slog.Any("response", sslDomainCertDeploySitesV2Resp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'v2.ssldomain.CertDeploySites': %w", err)
	}

	return nil
}

func createSDKClient(serverUrl, apiKey string, skipTlsVerify bool) (*btpanelsdk.Client, error) {
	client, err := btpanelsdk.NewClient(serverUrl,
		btpanelsdk.WithApiKey(apiKey),
	)
	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, nil
}
