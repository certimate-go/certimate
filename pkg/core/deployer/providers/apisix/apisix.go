package apisix

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"

	apisix "github.com/holubovskyi/apisix-client-go"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xhttp "github.com/certimate-go/certimate/pkg/utils/http"
)

type DeployerConfig struct {
	// APISIX 服务地址。
	ServerUrl string `json:"serverUrl"`
	// APISIX Admin API Key。
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
	sdkClient *apisix.ApiClient
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
	// 根据部署资源类型决定部署方式
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

	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return err
	}

	// 更新 SSL 证书
	// REF: https://apisix.apache.org/zh/docs/apisix/admin-api/#ssl
	updateSSLCertificateReq := &apisix.SSLCertificate{
		ID:          lo.ToPtr(d.config.CertificateId),
		Certificate: lo.ToPtr(certPEM),
		PrivateKey:  lo.ToPtr(privkeyPEM),
		SNIs:        lo.ToPtr(certX509.DNSNames),
		Type:        lo.ToPtr("server"),
		Status:      lo.ToPtr(int64(1)),
	}
	updateSSLCertificateResp, err := d.sdkClient.UpdateSslCertificate(d.config.CertificateId, *updateSSLCertificateReq)
	d.logger.Debug("sdk request 'apisix.UpdateSslCertificate'", slog.Any("request", updateSSLCertificateReq), slog.Any("response", updateSSLCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'apisix.UpdateSslCertificate': %w", err)
	}

	return nil
}

func createSDKClient(serverUrl, apiKey string, skipTlsVerify bool) (*apisix.ApiClient, error) {
	client, err := apisix.NewClient(&serverUrl, &apiKey)
	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		transport := xhttp.NewDefaultTransport()
		transport.DisableKeepAlives = true
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		if client.HTTPClient.Transport == nil {
			client.HTTPClient.Transport = transport
		} else {
			transport := client.HTTPClient.Transport.(*apisix.AddHeadersRoundtripper)
			transport.Nested = transport
			client.HTTPClient.Transport = transport
		}
	}

	return client, nil
}
