package netlify

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	netlifysdk "github.com/certimate-go/certimate/pkg/sdk3rd/netlify"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type DeployerConfig struct {
	// netlify API Token。
	ApiToken string `json:"apiToken"`
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// netlify 网站 ID。
	// 部署资源类型为 [RESOURCE_TYPE_WEBSITE] 时必填。
	SiteId string `json:"siteId,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *netlifysdk.Client
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
	// 根据部署资源类型决定部署方式
	switch d.config.ResourceType {
	case RESOURCE_TYPE_WEBSITE:
		if err := d.deployToWebsite(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported resource type '%s'", d.config.ResourceType)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) deployToWebsite(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.SiteId == "" {
		return errors.New("config `siteId` is required")
	}

	// 提取服务器证书和中间证书
	serverCertPEM, intermediaCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return fmt.Errorf("failed to extract certs: %w", err)
	}

	// 上传网站证书
	// REF: https://open-api.netlify.com/#tag/sniCertificate/operation/provisionSiteTLSCertificate
	provisionSiteTLSCertificateReq := &netlifysdk.ProvisionSiteTLSCertificateParams{
		Certificate:    serverCertPEM,
		CACertificates: intermediaCertPEM,
		Key:            privkeyPEM,
	}
	provisionSiteTLSCertificateResp, err := d.sdkClient.ProvisionSiteTLSCertificate(d.config.SiteId, provisionSiteTLSCertificateReq)
	d.logger.Debug("sdk request 'netlify.provisionSiteTLSCertificate'", slog.String("siteId", d.config.SiteId), slog.Any("request", provisionSiteTLSCertificateReq), slog.Any("response", provisionSiteTLSCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'netlify.provisionSiteTLSCertificate': %w", err)
	}

	return nil
}

func createSDKClient(apiToken string) (*netlifysdk.Client, error) {
	return netlifysdk.NewClient(apiToken)
}
