package baotapanelgoconsole

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	btsdk "github.com/certimate-go/certimate/pkg/sdk3rd/btpanelgo"
)

type SSLDeployerProviderConfig struct {
	// 宝塔面板服务地址。
	ServerUrl string `json:"serverUrl"`
	// 宝塔面板接口密钥。
	ApiKey string `json:"apiKey"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
}

type SSLDeployerProvider struct {
	config    *SSLDeployerProviderConfig
	logger    *slog.Logger
	sdkClient *btsdk.Client
}

var _ core.SSLDeployer = (*SSLDeployerProvider)(nil)

func NewSSLDeployerProvider(config *SSLDeployerProviderConfig) (*SSLDeployerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl deployer provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.ApiKey, config.AllowInsecureConnections)
	if err != nil {
		return nil, fmt.Errorf("could not create sdk client: %w", err)
	}

	return &SSLDeployerProvider{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (d *SSLDeployerProvider) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}
}

func (d *SSLDeployerProvider) Deploy(ctx context.Context, certPEM string, privkeyPEM string) (*core.SSLDeployResult, error) {
	// 设置面板 SSL 证书
	configSetPanelSSLReq := &btsdk.ConfigSetPanelSSLRequest{
		SSLStatus: lo.ToPtr(int32(1)),
		SSLKey:    lo.ToPtr(privkeyPEM),
		SSLPem:    lo.ToPtr(certPEM),
	}
	configSetPanelSSLResp, err := d.sdkClient.ConfigSetPanelSSL(configSetPanelSSLReq)
	d.logger.Debug("sdk request 'bt.ConfigSetPanelSSL'", slog.Any("request", configSetPanelSSLReq), slog.Any("response", configSetPanelSSLResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'bt.ConfigSetPanelSSL': %w", err)
	}

	return &core.SSLDeployResult{}, nil
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
