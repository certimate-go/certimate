package ratpanel

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"time"

	"github.com/certimate-go/certimate/pkg/core"
	ratpanelsdk "github.com/certimate-go/certimate/pkg/sdk3rd/ratpanel"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// 耗子面板服务地址。
	ServerUrl string `json:"serverUrl"`
	// 耗子面板访问令牌 ID。
	AccessTokenId int64 `json:"accessTokenId"`
	// 耗子面板访问令牌。
	AccessToken string `json:"accessToken"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 部署目标。
	DeployTarget string `json:"deployTarget"`
	// 网站名称。
	// 部署目标为 [DEPLOY_TARGET_WEBSITE] 时必填。
	SiteNames []string `json:"siteNames,omitempty"`
	// 证书 ID。
	// 部署目标为 [DEPLOY_TARGET_CERTIFICATE] 时必填。
	CertificateId int64 `json:"certificateId,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *ratpanelsdk.Client
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.AccessTokenId, config.AccessToken, config.AllowInsecureConnections)
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
		if err := d.deployToWebsite(ctx, certPEM, privkeyPEM); err != nil {
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

func (d *Deployer) deployToWebsite(ctx context.Context, certPEM, privkeyPEM string) error {
	if len(d.config.SiteNames) == 0 {
		return fmt.Errorf("config `siteNames` is required")
	}

	// 批量更新站点证书
	if err := xloop.ForRangeAllWithContext(ctx, d.config.SiteNames, func(ctx context.Context, siteName string, i int) error {
		if i > 0 {
			if err := xwait.DelayWithContext(ctx, 3*time.Second); err != nil {
				return err
			}
		}

		return d.updateSiteCertificate(ctx, siteName, certPEM, privkeyPEM)
	}); err != nil {
		return err
	}

	return nil
}

func (d *Deployer) deployToCertificate(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.CertificateId == 0 {
		return fmt.Errorf("config `certificateId` is required")
	}

	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return err
	}

	// 更新 SSL 证书
	certUpdateReq := &ratpanelsdk.CertUpdateRequest{
		CertId:      d.config.CertificateId,
		Type:        "upload",
		Domains:     certX509.DNSNames,
		Certificate: certPEM,
		PrivateKey:  privkeyPEM,
	}
	certUpdateResp, err := d.sdkClient.CertUpdateWithContext(ctx, certUpdateReq)
	d.logger.Debug("sdk request 'CertUpdate'", slog.Any("request", certUpdateReq), slog.Any("response", certUpdateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'CertUpdate': %w", err)
	}

	return nil
}

func (d *Deployer) updateSiteCertificate(ctx context.Context, siteName string, certPEM, privkeyPEM string) error {
	// 设置站点 SSL 证书
	setWebsiteCertReq := &ratpanelsdk.SetWebsiteCertRequest{
		SiteName:    siteName,
		Certificate: certPEM,
		PrivateKey:  privkeyPEM,
	}
	setWebsiteCertResp, err := d.sdkClient.SetWebsiteCertWithContext(ctx, setWebsiteCertReq)
	d.logger.Debug("sdk request 'SetWebsiteCert'", slog.Any("request", setWebsiteCertReq), slog.Any("response", setWebsiteCertResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'SetWebsiteCert': %w", err)
	}

	return nil
}

func createSDKClient(serverUrl string, accessTokenId int64, accessToken string, skipTlsVerify bool) (*ratpanelsdk.Client, error) {
	client, err := ratpanelsdk.NewClient(serverUrl,
		ratpanelsdk.WithAccessToken(accessTokenId, accessToken),
	)
	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, nil
}
