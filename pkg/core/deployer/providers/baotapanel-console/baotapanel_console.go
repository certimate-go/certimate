package baotapanelconsole

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/pkg/core"
	btpanelsdk "github.com/certimate-go/certimate/pkg/sdk3rd/btpanel"
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
	// 是否自动重启。
	AutoRestart bool `json:"autoRestart"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *btpanelsdk.Client
}

var _ Provider = (*Deployer)(nil)

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
	// 设置面板 SSL 证书
	configSavePanelSSLReq := &btpanelsdk.ConfigSavePanelSSLRequest{
		PrivateKey:  privkeyPEM,
		Certificate: certPEM,
	}
	configSavePanelSSLResp, err := d.sdkClient.ConfigSavePanelSSLWithContext(ctx, configSavePanelSSLReq)
	d.logger.Debug("sdk request 'config.SavePanelSSL'", slog.Any("request", configSavePanelSSLReq), slog.Any("response", configSavePanelSSLResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'config.SavePanelSSL': %w", err)
	}

	if d.config.AutoRestart {
		// 重启面板（无需关心响应，因为宝塔重启时会断开连接产生 error）
		systemServiceAdminReq := &btpanelsdk.SystemServiceAdminRequest{
			Name: "nginx",
			Type: "restart",
		}
		systemServiceAdminResp, _ := d.sdkClient.SystemServiceAdminWithContext(ctx, systemServiceAdminReq)
		d.logger.Debug("sdk request 'system.ServiceAdmin'", slog.Any("request", systemServiceAdminReq), slog.Any("response", systemServiceAdminResp))
	}

	return &DeployResult{}, nil
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
