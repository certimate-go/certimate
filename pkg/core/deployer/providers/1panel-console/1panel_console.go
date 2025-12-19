package onepanelconsole

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	onepanelsdk "github.com/certimate-go/certimate/pkg/sdk3rd/1panel"
	onepanelsdk2 "github.com/certimate-go/certimate/pkg/sdk3rd/1panel/v2"
)

type DeployerConfig struct {
	// 1Panel 服务地址。
	ServerUrl string `json:"serverUrl"`
	// 1Panel 版本。
	// 可取值 "v1"、"v2"。
	ApiVersion string `json:"apiVersion"`
	// 1Panel 接口密钥。
	ApiKey string `json:"apiKey"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 是否自动重启。
	AutoRestart bool `json:"autoRestart"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient any
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.ApiVersion, config.ApiKey, config.AllowInsecureConnections)
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
	switch sdkClient := d.sdkClient.(type) {
	case *onepanelsdk.Client:
		{
			settingsSSLUpdateReq := &onepanelsdk.SettingsSSLUpdateRequest{
				Cert:        certPEM,
				Key:         privkeyPEM,
				SSL:         "enable",
				SSLType:     "import-paste",
				AutoRestart: strconv.FormatBool(d.config.AutoRestart),
			}
			settingsSSLUpdateResp, err := sdkClient.SettingsSSLUpdate(settingsSSLUpdateReq)
			d.logger.Debug("sdk request '1panel.SettingsSSLUpdate'", slog.Any("request", settingsSSLUpdateReq), slog.Any("response", settingsSSLUpdateResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request '1panel.SettingsSSLUpdate': %w", err)
			}
		}

	case *onepanelsdk2.Client:
		{
			coreSettingsSSLUpdateReq := &onepanelsdk2.CoreSettingsSSLUpdateRequest{
				Cert:        certPEM,
				Key:         privkeyPEM,
				SSL:         "Enable",
				SSLType:     "import-paste",
				AutoRestart: strconv.FormatBool(d.config.AutoRestart),
			}
			coreSettingsSSLUpdateResp, err := sdkClient.CoreSettingsSSLUpdate(coreSettingsSSLUpdateReq)
			d.logger.Debug("sdk request '1panel.CoreSettingsSSLUpdate'", slog.Any("request", coreSettingsSSLUpdateReq), slog.Any("response", coreSettingsSSLUpdateResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request '1panel.CoreSettingsSSLUpdate': %w", err)
			}
		}

	default:
		panic("unreachable")
	}

	return &deployer.DeployResult{}, nil
}

const (
	sdkVersionV1 = "v1"
	sdkVersionV2 = "v2"
)

func createSDKClient(serverUrl, apiVersion, apiKey string, skipTlsVerify bool) (any, error) {
	if apiVersion == sdkVersionV1 {
		client, err := onepanelsdk.NewClient(serverUrl, apiKey)
		if err != nil {
			return nil, err
		}

		if skipTlsVerify {
			client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
		}

		return client, nil
	} else if apiVersion == sdkVersionV2 {
		client, err := onepanelsdk2.NewClient(serverUrl, apiKey)
		if err != nil {
			return nil, err
		}

		if skipTlsVerify {
			client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
		}

		return client, nil
	}

	return nil, errors.New("1panel: invalid api version")
}
