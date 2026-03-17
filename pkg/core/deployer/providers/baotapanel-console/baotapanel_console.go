package baotapanelconsole

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"

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
	// 是否自动重启。
	AutoRestart bool `json:"autoRestart"`
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
	// 设置面板 SSL 证书
	configSavePanelSSLReq := &btsdk.ConfigSavePanelSSLRequest{
		PrivateKey:  privkeyPEM,
		Certificate: certPEM,
	}
	configSavePanelSSLResp, err := d.sdkClient.ConfigSavePanelSSLWithContext(ctx, configSavePanelSSLReq)
	d.logger.Debug("sdk request 'bt.ConfigSavePanelSSL'", slog.Any("request", configSavePanelSSLReq), slog.Any("response", configSavePanelSSLResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'bt.ConfigSavePanelSSL': %w", err)
	}

	if d.config.AutoRestart {
		// 重启面板（无需关心响应，因为宝塔重启时会断开连接产生 error）
		systemServiceAdminReq := &btsdk.SystemServiceAdminRequest{
			Name: "nginx",
			Type: "restart",
		}
		systemServiceAdminResp, _ := d.sdkClient.SystemServiceAdminWithContext(ctx, systemServiceAdminReq)
		d.logger.Debug("sdk request 'bt.SystemServiceAdmin'", slog.Any("request", systemServiceAdminReq), slog.Any("response", systemServiceAdminResp))
	}

	return &deployer.DeployResult{}, nil
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
