package cachefly

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	cacheflysdk "github.com/certimate-go/certimate/pkg/sdk3rd/cachefly"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// CacheFly API Token。
	ApiToken string `json:"apiToken"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *cacheflysdk.Client
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
	// 上传证书
	// REF: https://api.cachefly.com/api/v2/docs/api/#tag/Certificates/operation/post-certificates
	createCertificateReq := &cacheflysdk.CreateCertificateRequest{
		Certificate:    lo.ToPtr(certPEM),
		CertificateKey: lo.ToPtr(privkeyPEM),
	}
	createCertificateResp, err := d.sdkClient.CreateCertificateWithContext(ctx, createCertificateReq)
	d.logger.Debug("sdk request 'CreateCertificate'", slog.Any("request", createCertificateReq), slog.Any("response", createCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'CreateCertificate': %w", err)
	}

	return &DeployResult{}, nil
}

func createSDKClient(apiToken string) (*cacheflysdk.Client, error) {
	client, err := cacheflysdk.NewClient(
		cacheflysdk.WithApiToken(apiToken),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
