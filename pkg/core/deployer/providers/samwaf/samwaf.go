package samwaf

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	samwafsdk "github.com/certimate-go/certimate/pkg/sdk3rd/samwaf"
)

type DeployerConfig struct {
	// SamWAF 服务地址。
	ServerUrl string `json:"serverUrl"`
	// SamWAF API Key。
	ApiKey string `json:"apiKey"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// 证书 ID。
	// 部署资源类型为 [RESOURCE_TYPE_CERTIFICATE] 时必填。
	CertificateId string `json:"certificateId,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *samwafsdk.Client
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.ApiKey, config.AllowInsecureConnections)
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
	// 根据部署资源类型决定部署方式``
	switch d.config.ResourceType {
	case RESOURCE_TYPE_CERTIFICATE:
		if err := d.deployToCertificate(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported resource type '%s'", d.config.ResourceType)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) deployToCertificate(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.CertificateId == "" {
		return errors.New("config `certificateId` is required")
	}

	// 获取 SSL 证书 详情
	// REF: https://doc.samwaf.com/api/
	sslConfigDetailResp, err := d.sdkClient.SslConfigDetailWithContext(ctx, d.config.CertificateId)
	d.logger.Debug("sdk request 'samwaf.SslConfigDetail'", slog.Any("request.sslId", d.config.CertificateId), slog.Any("response", sslConfigDetailResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'samwaf.SslConfigDetail': %w", err)
	} else if sslConfigDetailResp.Data == nil || sslConfigDetailResp.Data.Id == "" {
		return fmt.Errorf("could not find ssl config: '%s'", d.config.CertificateId)
	}

	// 编辑 SSL 证书
	// REF: https://doc.samwaf.com/api/
	sslConfigEditReq := &samwafsdk.SslConfigEditRequest{
		Id:          d.config.CertificateId,
		CertContent: certPEM,
		KeyContent:  privkeyPEM,
	}
	sslConfigEditResp, err := d.sdkClient.SslConfigEditWithContext(ctx, sslConfigEditReq)
	d.logger.Debug("sdk request 'samwaf.SslConfigEdit'", slog.Any("request", sslConfigEditReq), slog.Any("response", sslConfigEditResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'samwaf.SslConfigEdit': %w", err)
	}

	return nil
}

func createSDKClient(serverUrl, apiKey string, skipTlsVerify bool) (*samwafsdk.Client, error) {
	client, err := samwafsdk.NewClient(serverUrl, apiKey)
	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, nil
}
