package cloudflaressl

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"

	cf "github.com/cloudflare/cloudflare-go/v7"
	cfcertificates "github.com/cloudflare/cloudflare-go/v7/custom_certificates"
	cfhostnames "github.com/cloudflare/cloudflare-go/v7/custom_hostnames"
	cfoption "github.com/cloudflare/cloudflare-go/v7/option"

	"github.com/certimate-go/certimate/pkg/core"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// Cloudflare API Token。
	ApiToken string `json:"apiToken"`
	// Cloudflare 环境。
	// 选填。
	// 零值时默认值 "production"。
	Environment string `json:"environment,omitempty"`
	// Cloudflare DNS 区域 ID。
	ZoneId string `json:"zoneId"`
	// Cloudflare 证书 ID。
	// 选填。零值时表示新建证书；否则表示更新证书。
	CertificateId string `json:"certificateId,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *cfcertificates.CustomCertificateService
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
	if d.config.ZoneId == "" {
		return nil, fmt.Errorf("config `zoneId` is required")
	}

	if d.config.CertificateId == "" {
		// 新建自定义证书
		// REF: https://developers.cf.com/api/resources/custom_certificates/methods/create
		customCertificateNewReq := cfcertificates.CustomCertificateNewParams{
			ZoneID:       cf.F(d.config.ZoneId),
			Certificate:  cf.F(certPEM),
			PrivateKey:   cf.F(privkeyPEM),
			BundleMethod: cf.F(cfhostnames.BundleMethodUbiquitous),
			Deploy:       cf.F(cfcertificates.CustomCertificateNewParamsDeploy(cmp.Or(d.config.Environment, "production"))),
		}
		customCertificateNewResp, err := d.sdkClient.New(ctx, customCertificateNewReq)
		d.logger.Debug("sdk request 'CustomCertificates.New'", slog.Any("request", customCertificateNewReq), slog.Any("response", customCertificateNewResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'CustomCertificates.New': %w", err)
		}
	} else {
		// 编辑自定义证书
		// REF: https://developers.cloudflare.com/api/resources/custom_certificates/methods/edit
		customCertificateEditReq := cfcertificates.CustomCertificateEditParams{
			ZoneID:       cf.F(d.config.ZoneId),
			Certificate:  cf.F(certPEM),
			PrivateKey:   cf.F(privkeyPEM),
			BundleMethod: cf.F(cfhostnames.BundleMethodUbiquitous),
			Deploy:       cf.F(cfcertificates.CustomCertificateEditParamsDeploy(cmp.Or(d.config.Environment, "production"))),
		}
		customCertificateEditResp, err := d.sdkClient.Edit(ctx, d.config.CertificateId, customCertificateEditReq)
		d.logger.Debug("sdk request 'CustomCertificates.Edit'", slog.Any("request", customCertificateEditReq), slog.Any("response", customCertificateEditResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'CustomCertificates.Edit': %w", err)
		}
	}

	return &DeployResult{}, nil
}

func createSDKClient(apiToken string) (*cfcertificates.CustomCertificateService, error) {
	if apiToken == "" {
		return nil, fmt.Errorf("cloudflare: invalid api token")
	}

	opts := append(cf.DefaultClientOptions(), cfoption.WithAPIToken(apiToken))

	srv := cfcertificates.NewCustomCertificateService(opts...)

	return srv, nil
}
