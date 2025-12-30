package ratpanel

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	ratpanelsdk "github.com/certimate-go/certimate/pkg/sdk3rd/ratpanel"
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
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// 网站名称。
	// 部署资源类型为 [RESOURCE_TYPE_WEBSITE] 时必填。
	SiteNames []string `json:"siteNames,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *ratpanelsdk.Client
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	// 根据部署资源类型决定部署方式
	switch d.config.ResourceType {
	case RESOURCE_TYPE_WEBSITE:
		if err := d.deployToWebsite(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported resource type '%s'", d.config.ResourceType)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) deployToWebsite(ctx context.Context, certPEM, privkeyPEM string) error {
	if len(d.config.SiteNames) == 0 {
		return errors.New("config `siteNames` is required")
	}

	// 遍历更新站点证书
	var errs []error
	for _, siteName := range d.config.SiteNames {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := d.updateSiteCertificate(ctx, siteName, certPEM, privkeyPEM); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
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
	setWebsiteCertResp, err := d.sdkClient.SetWebsiteCert(setWebsiteCertReq)
	d.logger.Debug("sdk request 'ratpanel.SetWebsiteCert'", slog.Any("request", setWebsiteCertReq), slog.Any("response", setWebsiteCertResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'ratpanel.SetWebsiteCert': %w", err)
	}

	return nil
}

func createSDKClient(serverUrl string, accessTokenId int64, accessToken string, skipTlsVerify bool) (*ratpanelsdk.Client, error) {
	client, err := ratpanelsdk.NewClient(serverUrl, accessTokenId, accessToken)
	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, nil
}
