package flyio

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/pkg/core"
	flyiosdk "github.com/certimate-go/certimate/pkg/sdk3rd/flyio"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// Fly.io API Token。
	ApiToken string `json:"apiToken"`
	// Fly.io 应用名称。
	AppName string `json:"appName"`
	// 自定义域名（支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *flyiosdk.Client
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*DeployResult, error) {
	if d.config.AppName == "" {
		return nil, fmt.Errorf("config `appName` is required")
	}
	if d.config.Domain == "" {
		return nil, fmt.Errorf("config `domain` is required")
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
	d.logger.Debug("sdk request 'ImportCustomCertificate'", slog.Any("request", importCustomCertificateReq), slog.Any("response", importCustomCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'ImportCustomCertificate': %w", err)
	}

	return &DeployResult{}, nil
}

func createSDKClient(apiToken string) (*flyiosdk.Client, error) {
	client, err := flyiosdk.NewClient(
		flyiosdk.WithApiToken(apiToken),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
