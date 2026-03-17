package baotapanelconsole

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	btwafsdk "github.com/certimate-go/certimate/pkg/sdk3rd/btwaf"
)

type DeployerConfig struct {
	// 堡塔云 WAF 服务地址。
	ServerUrl string `json:"serverUrl"`
	// 堡塔云 WAF 接口密钥。
	ApiKey string `json:"apiKey"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *btwafsdk.Client
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
	// 设置面板 SSL
	configSetCertReq := &btwafsdk.ConfigSetCertRequest{
		CertContent: lo.ToPtr(certPEM),
		KeyContent:  lo.ToPtr(privkeyPEM),
	}
	configSetCertResp, err := d.sdkClient.ConfigSetCertWithContext(ctx, configSetCertReq)
	d.logger.Debug("sdk request 'bt.ConfigSetCert'", slog.Any("request", configSetCertReq), slog.Any("response", configSetCertResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'bt.ConfigSetCert': %w", err)
	}

	return &deployer.DeployResult{}, nil
}

func createSDKClient(serverUrl, apiKey string, skipTlsVerify bool) (*btwafsdk.Client, error) {
	client, err := btwafsdk.NewClient(serverUrl, apiKey)
	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, nil
}
