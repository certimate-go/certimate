package flyio

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	flyiosdk "github.com/certimate-go/certimate/pkg/sdk3rd/flyio"
)

type DeployerConfig struct {
	// Fly.io API Token。
	ApiToken string `json:"apiToken"`
	// Fly.io 应用名称。
	AppName string `json:"appName"`
	// 自定义域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *flyiosdk.Client
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ApiToken)
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
	if d.config.AppName == "" {
		return nil, errors.New("config `appName` is required")
	}
	if d.config.Domain == "" {
		return nil, errors.New("config `domain` is required")
	}

	// 导入自定义证书
	// REF: https://fly.io/docs/machines/api/certificates-resource/#import-custom-certificate
	importCustomCertificateReq := &flyiosdk.ImportCustomCertificateRequest{
		AppName:    d.config.AppName,
		Hostname:   d.config.Domain,
		Fullchain:  certPEM,
		PrivateKey: privkeyPEM,
	}
	importCustomCertificateResp, err := d.sdkClient.ImportCustomCertificateWithContext(ctx, importCustomCertificateReq)
	d.logger.Debug("sdk request 'flyio.ImportCustomCertificate'", slog.Any("request", importCustomCertificateReq), slog.Any("response", importCustomCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'flyio.ImportCustomCertificate': %w", err)
	}

	return &deployer.DeployResult{}, nil
}

func createSDKClient(apiToken string) (*flyiosdk.Client, error) {
	return flyiosdk.NewClient(apiToken)
}
