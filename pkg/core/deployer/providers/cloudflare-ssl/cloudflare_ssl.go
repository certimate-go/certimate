package cloudflaressl

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	cloudflaresdk "github.com/certimate-go/certimate/pkg/sdk3rd/cloudflare"
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
	sdkClient *cloudflaresdk.Client
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
		customCertificateCreateReq := &cloudflaresdk.CustomCertificateCreateRequest{
			ZoneId:       d.config.ZoneId,
			Certificate:  lo.ToPtr(certPEM),
			PrivateKey:   lo.ToPtr(privkeyPEM),
			BundleMethod: lo.ToPtr("ubiquitous"),
			Deploy:       lo.ToPtr(cmp.Or(d.config.Environment, "production")),
		}
		customCertificateCreateResp, err := d.sdkClient.CustomCertificateCreateWithContext(ctx, customCertificateCreateReq)
		d.logger.Debug("sdk request 'CustomCertificates.Create'", slog.Any("request", customCertificateCreateReq), slog.Any("response", customCertificateCreateResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'CustomCertificates.Create': %w", err)
		}
	} else {
		// 编辑自定义证书
		// REF: https://developers.cloudflare.com/api/resources/custom_certificates/methods/edit
		customCertificateEditReq := &cloudflaresdk.CustomCertificateEditRequest{
			ZoneId:        d.config.ZoneId,
			CertificateId: d.config.CertificateId,
			Certificate:   lo.ToPtr(certPEM),
			PrivateKey:    lo.ToPtr(privkeyPEM),
			BundleMethod:  lo.ToPtr("ubiquitous"),
			Deploy:        lo.ToPtr(cmp.Or(d.config.Environment, "production")),
		}
		customCertificateEditResp, err := d.sdkClient.CustomCertificateEditWithContext(ctx, customCertificateEditReq)
		d.logger.Debug("sdk request 'CustomCertificates.Edit'", slog.Any("request", customCertificateEditReq), slog.Any("response", customCertificateEditResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'CustomCertificates.Edit': %w", err)
		}
	}

	return &DeployResult{}, nil
}

func createSDKClient(apiToken string) (*cloudflaresdk.Client, error) {
	client, err := cloudflaresdk.NewClient(
		cloudflaresdk.WithApiToken(apiToken),
	)
	if err != nil {
		return nil, err
	}

	return client, err
}
