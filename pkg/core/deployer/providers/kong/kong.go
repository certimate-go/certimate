package kong

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/kong/go-kong/kong"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xhttp "github.com/certimate-go/certimate/pkg/utils/http"
)

type DeployerConfig struct {
	// Kong 服务地址。
	ServerUrl string `json:"serverUrl"`
	// Kong Admin API Token。
	ApiToken string `json:"apiToken,omitempty"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// 工作空间。
	// 选填。
	Workspace string `json:"workspace,omitempty"`
	// 证书 ID。
	// 部署资源类型为 [RESOURCE_TYPE_CERTIFICATE] 时必填。
	CertificateId string `json:"certificateId,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *kong.Client
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
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

	// 更新证书
	// REF: https://developer.konghq.com/api/gateway/admin-ee/3.10/#/operations/upsert-certificate
	// REF: https://developer.konghq.com/api/gateway/admin-ee/3.10/#/operations/upsert-certificate-in-workspace
	updateCertificateReq := &kong.Certificate{
		ID:   kong.String(d.config.CertificateId),
		Cert: kong.String(certPEM),
		Key:  kong.String(privkeyPEM),
		SNIs: kong.StringSlice(certX509.DNSNames...),
	}
	updateCertificateResp, err := d.sdkClient.Certificates.Update(ctx, updateCertificateReq)
	d.logger.Debug("sdk request 'kong.UpdateCertificate'", slog.String("sslId", d.config.CertificateId), slog.Any("request", updateCertificateReq), slog.Any("response", updateCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'kong.UpdateCertificate': %w", err)
	}

	return nil
}

func createSDKClient(serverUrl, workspace, apiToken string, skipTlsVerify bool) (*kong.Client, error) {
	httpClient := &http.Client{
		Transport: xhttp.NewDefaultTransport(),
		Timeout:   http.DefaultClient.Timeout,
	}
	if skipTlsVerify {
		transport := xhttp.NewDefaultTransport()
		transport.DisableKeepAlives = true
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient.Transport = transport
	} else {
		httpClient.Transport = http.DefaultTransport
	}

	httpHeaders := http.Header{}
	if apiToken != "" {
		httpHeaders.Set("Kong-Admin-Token", apiToken)
	}

	client, err := kong.NewClient(kong.String(serverUrl), kong.HTTPClientWithHeaders(httpClient, httpHeaders))
	if err != nil {
		return nil, err
	}

	if workspace != "" {
		client.SetWorkspace(workspace)
	}

	return client, nil
}
