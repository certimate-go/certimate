package cpanel

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	cpanelsdk "github.com/certimate-go/certimate/pkg/sdk3rd/cpanel"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type DeployerConfig struct {
	// cPanel 服务地址。
	ServerUrl string `json:"serverUrl"`
	// cPanel 用户名。
	Username string `json:"username"`
	// cPanel 接口密钥。
	ApiToken string `json:"apiToken"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// 网站域名（不支持泛域名）。
	// 部署资源类型为 [RESOURCE_TYPE_WEBSITE] 时必填。
	Domain string `json:"domain,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *cpanelsdk.Client
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.Username, config.ApiToken, config.AllowInsecureConnections)
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
	if d.config.Domain == "" {
		return errors.New("config `domain` is required")
	}

	// 提取服务器证书和中间证书
	serverCertPEM, intermediaCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return fmt.Errorf("failed to extract certs: %w", err)
	}

	// 安装 SSL 证书
	// REF: https://api.docs.cpanel.net/openapi/cpanel/operation/install_ssl/
	sslInstallSSLReq := &cpanelsdk.SSLInstallSSLRequest{
		Domain:   lo.ToPtr(d.config.Domain),
		Cert:     lo.ToPtr(serverCertPEM),
		Key:      lo.ToPtr(privkeyPEM),
		CABundle: lo.ToPtr(intermediaCertPEM),
	}
	sslInstallSSLResp, err := d.sdkClient.SSLInstallSSL(sslInstallSSLReq)
	d.logger.Debug("sdk request 'SSL.install_ssl'", slog.Any("request", sslInstallSSLReq), slog.Any("response", sslInstallSSLResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'SSL.install_ssl': %w", err)
	}

	return nil
}

func createSDKClient(serverUrl, username, apiToken string, skipTlsVerify bool) (*cpanelsdk.Client, error) {
	client, err := cpanelsdk.NewClient(serverUrl, username, apiToken)
	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, nil
}
