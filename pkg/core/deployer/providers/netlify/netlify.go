package netlify

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/pkg/core"
	netlifysdk "github.com/certimate-go/certimate/pkg/sdk3rd/netlify"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// netlify API Token。
	ApiToken string `json:"apiToken"`
	// 部署目标。
	DeployTarget string `json:"deployTarget"`
	// netlify 网站 ID。
	// 部署目标为 [DEPLOY_TARGET_WEBSITE] 时必填。
	SiteId string `json:"siteId,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *netlifysdk.Client
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
	// 根据部署目标决定业务流程
	switch d.config.DeployTarget {
	case DEPLOY_TARGET_WEBSITE:
		if err := d.deployToWebsite(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported deploy target '%s'", d.config.DeployTarget)
	}

	return &DeployResult{}, nil
}

func (d *Deployer) deployToWebsite(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.SiteId == "" {
		return fmt.Errorf("config `siteId` is required")
	}

	// 提取服务器证书和中间证书
	serverCertPEM, issuerCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return fmt.Errorf("failed to extract certs: %w", err)
	}

	// 上传网站证书
	// REF: https://open-api.netlify.com/#tag/sniCertificate/operation/provisionSiteTLSCertificate
	provisionSiteTLSCertificateReq := &netlifysdk.ProvisionSiteTLSCertificateRequest{
		Certificate:    serverCertPEM,
		CACertificates: issuerCertPEM,
		Key:            privkeyPEM,
	}
	provisionSiteTLSCertificateResp, err := d.sdkClient.ProvisionSiteTLSCertificateWithContext(ctx, d.config.SiteId, provisionSiteTLSCertificateReq)
	d.logger.Debug("sdk request 'ProvisionSiteTLSCertificate'", slog.String("params.siteId", d.config.SiteId), slog.Any("request", provisionSiteTLSCertificateReq), slog.Any("response", provisionSiteTLSCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'ProvisionSiteTLSCertificate': %w", err)
	}

	return nil
}

func createSDKClient(apiToken string) (*netlifysdk.Client, error) {
	client, err := netlifysdk.NewClient(
		netlifysdk.WithApiToken(apiToken),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
