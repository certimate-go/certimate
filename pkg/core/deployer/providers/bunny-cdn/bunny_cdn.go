package bunnycdn

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	bunnysdk "github.com/certimate-go/certimate/pkg/sdk3rd/bunny"
)

type DeployerConfig struct {
	// Bunny API Key。
	ApiKey string `json:"apiKey"`
	// Bunny Pull Zone ID。
	PullZoneId string `json:"pullZoneId"`
	// Bunny CDN Hostname。
	Hostname string `json:"hostname"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *bunnysdk.Client
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ApiKey)
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
	if d.config.PullZoneId == "" {
		return nil, fmt.Errorf("config `pullZoneId` is required")
	}
	if d.config.Hostname == "" {
		return nil, fmt.Errorf("config `hostname` is required")
	}

	// 上传证书
	createCertificateReq := &bunnysdk.AddCustomCertificateRequest{
		Hostname:       d.config.Hostname,
		Certificate:    base64.StdEncoding.EncodeToString([]byte(certPEM)),
		CertificateKey: base64.StdEncoding.EncodeToString([]byte(privkeyPEM)),
	}
	err := d.sdkClient.AddCustomCertificateWithContext(ctx, d.config.PullZoneId, createCertificateReq)
	d.logger.Debug("sdk request 'bunny.AddCustomCertificate'", slog.Any("request", createCertificateReq))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'bunny.AddCustomCertificate': %w", err)
	}

	return &deployer.DeployResult{}, nil
}

func createSDKClient(apiKey string) (*bunnysdk.Client, error) {
	return bunnysdk.NewClient(apiKey)
}
