package baotapanel

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	btsdk "github.com/certimate-go/certimate/pkg/sdk3rd/btpanel"
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
	sdkClient *btsdk.Client
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	if len(d.config.SiteNames) == 0 {
		return nil, errors.New("config `siteNames` is required")
	}

	switch d.config.SiteType {
	case "any":
		{
			// 上传证书
			sslCertSaveCertReq := &btsdk.SSLCertSaveCertRequest{
				Certificate: certPEM,
				PrivateKey:  privkeyPEM,
			}
			sslCertSaveCertResp, err := d.sdkClient.SSLCertSaveCert(sslCertSaveCertReq)
			d.logger.Debug("sdk request 'bt.SSLCertSaveCert'", slog.Any("request", sslCertSaveCertReq), slog.Any("response", sslCertSaveCertResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request 'bt.SSLCertSaveCert': %w", err)
			}

			// 设置站点证书
			sslSetBatchCertToSiteReq := &btsdk.SSLSetBatchCertToSiteRequest{
				BatchInfo: lo.Map(d.config.SiteNames, func(siteName string, _ int) *btsdk.SSLSetBatchCertToSiteRequestBatchInfo {
					return &btsdk.SSLSetBatchCertToSiteRequestBatchInfo{
						SiteName: siteName,
						SSLHash:  sslCertSaveCertResp.SSLHash,
					}
				}),
			}
			sslSetBatchCertToSiteResp, err := d.sdkClient.SSLSetBatchCertToSite(sslSetBatchCertToSiteReq)
			d.logger.Debug("sdk request 'bt.SSLSetBatchCertToSite'", slog.Any("request", sslSetBatchCertToSiteReq), slog.Any("response", sslSetBatchCertToSiteResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request 'bt.SSLSetBatchCertToSite': %w", err)
			}
		}

	default:
		{
			// 遍历更新站点证书
			var errs []error
			for i, siteName := range d.config.SiteNames {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
					if err := d.updateSiteCertificate(ctx, siteName, certPEM, privkeyPEM); err != nil {
						errs = append(errs, err)
					}
					if i < len(d.config.SiteNames)-1 {
						time.Sleep(time.Second * 5)
					}
				}
			}
			if len(errs) > 0 {
				return nil, errors.Join(errs...)
			}
		}
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) updateSiteCertificate(ctx context.Context, siteName string, certPEM, privkeyPEM string) error {
	// 设置站点 SSL 证书
	siteSetSSLReq := &btsdk.SiteSetSSLRequest{
		SiteName:    siteName,
		Type:        "0",
		Certificate: certPEM,
		PrivateKey:  privkeyPEM,
	}
	siteSetSSLResp, err := d.sdkClient.SiteSetSSL(siteSetSSLReq)
	d.logger.Debug("sdk request 'bt.SiteSetSSL'", slog.Any("request", siteSetSSLReq), slog.Any("response", siteSetSSLResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'bt.SiteSetSSL': %w", err)
	}

	return nil
}

func createSDKClient(serverUrl, apiKey string, skipTlsVerify bool) (*btsdk.Client, error) {
	client, err := btsdk.NewClient(serverUrl, apiKey)
	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, nil
}
