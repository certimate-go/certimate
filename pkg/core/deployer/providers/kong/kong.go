package kong

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	kongsdk "github.com/certimate-go/certimate/pkg/sdk3rd/kong"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// Kong 服务地址。
	ServerUrl string `json:"serverUrl"`
	// Kong Admin API Token。
	ApiToken string `json:"apiToken,omitempty"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 部署目标。
	DeployTarget string `json:"deployTarget"`
	// 工作空间。
	// 选填。
	Workspace string `json:"workspace,omitempty"`
	// 证书 ID。
	// 部署目标为 [DEPLOY_TARGET_CERTIFICATE] 时必填。
	CertificateId string `json:"certificateId,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *kongsdk.Client
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.Workspace, config.ApiToken, config.AllowInsecureConnections)
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
	case DEPLOY_TARGET_CERTIFICATE:
		if err := d.deployToCertificate(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported deploy target '%s'", d.config.DeployTarget)
	}

	return &DeployResult{}, nil
}

func (d *Deployer) deployToCertificate(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.CertificateId == "" {
		return fmt.Errorf("config `certificateId` is required")
	}

	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return err
	}

	// 更新证书
	// REF: https://developer.konghq.com/api/gateway/admin-ee/3.10/#/operations/upsert-certificate
	// REF: https://developer.konghq.com/api/gateway/admin-ee/3.10/#/operations/upsert-certificate-in-workspace
	upsertCertificateReq := &kongsdk.UpsertCertificateRequest{
		Id:   lo.ToPtr(d.config.CertificateId),
		Cert: lo.ToPtr(certPEM),
		Key:  lo.ToPtr(privkeyPEM),
		SNIs: lo.ToSlicePtr(certX509.DNSNames),
	}
	upsertCertificateResp, err := d.sdkClient.UpsertCertificateWithContext(ctx, d.config.CertificateId, upsertCertificateReq)
	d.logger.Debug("sdk request 'UpsertCertificate'", slog.String("params.certificateId", d.config.CertificateId), slog.Any("request", upsertCertificateReq), slog.Any("response", upsertCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'UpsertCertificate': %w", err)
	}

	return nil
}

func createSDKClient(serverUrl, workspace, apiToken string, skipTlsVerify bool) (*kongsdk.Client, error) {
	client, err := kongsdk.NewClient(serverUrl,
		kongsdk.WithWorkspace(workspace),
		kongsdk.WithApiToken(apiToken),
	)
	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, err
}
