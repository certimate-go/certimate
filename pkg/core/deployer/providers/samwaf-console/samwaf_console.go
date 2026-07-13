package samwafconsole

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/pkg/core"
	samwafsdk "github.com/certimate-go/certimate/pkg/sdk3rd/samwaf"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// SamWAF 服务地址。
	ServerUrl string `json:"serverUrl"`
	// SamWAF API Key。
	ApiKey string `json:"apiKey"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 是否自动重启。
	AutoRestart bool `json:"autoRestart"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *samwafsdk.Client
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
	// 上传管理端 SSL 证书
	// REF: https://doc.samwaf.com/api/
	vipConfigUploadSslCertReq := &samwafsdk.VipConfigUploadSslCertRequest{
		CertContent: certPEM,
		KeyContent:  privkeyPEM,
	}
	vipConfigUploadSslCertResp, err := d.sdkClient.VipConfigUploadSslCertWithContext(ctx, vipConfigUploadSslCertReq)
	d.logger.Debug("sdk request 'vipconfig.UploadSslCert'", slog.Any("request", vipConfigUploadSslCertReq), slog.Any("response", vipConfigUploadSslCertResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'vipconfig.UploadSslCert': %w", err)
	}

	// 更新管理端 SSL 启用状态
	// REF: https://doc.samwaf.com/api/
	vipConfigUpdateSslEnableReq := &samwafsdk.VipConfigUpdateSslEnableRequest{
		SslEnable: true,
	}
	vipConfigUpdateSslEnableResp, err := d.sdkClient.VipConfigUpdateSslEnableWithContext(ctx, vipConfigUpdateSslEnableReq)
	d.logger.Debug("sdk request 'vipconfig.UpdateSslEnable'", slog.Any("request", vipConfigUpdateSslEnableReq), slog.Any("response", vipConfigUpdateSslEnableResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'vipconfig.UpdateSslEnable': %w", err)
	}

	if d.config.AutoRestart {
		// 重启管理端
		vipConfigRestartManagerResp, err := d.sdkClient.VipConfigRestartManagerWithContext(ctx)
		d.logger.Debug("sdk request 'vipconfig.RestartManager'", slog.Any("response", vipConfigRestartManagerResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'vipconfig.RestartManager': %w", err)
		}
	}

	return &DeployResult{}, nil
}

func createSDKClient(serverUrl, apiKey string, skipTlsVerify bool) (*samwafsdk.Client, error) {
	client, err := samwafsdk.NewClient(serverUrl,
		samwafsdk.WithApiKey(apiKey),
	)
	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, nil
}
