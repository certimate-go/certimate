package ratpanelconsole

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
	// 设置面板 SSL 证书
	setSettingCertReq := &ratpanelsdk.SetSettingCertRequest{
		Certificate: certPEM,
		PrivateKey:  privkeyPEM,
	}
	setSettingCertResp, err := d.sdkClient.SetSettingCertWithContext(ctx, setSettingCertReq)
	d.logger.Debug("sdk request 'ratpanel.SetSettingCert'", slog.Any("request", setSettingCertReq), slog.Any("response", setSettingCertResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'ratpanel.SetSettingCert': %w", err)
	}

	return &deployer.DeployResult{}, nil
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
