package proxmoxbs

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/pkg/core"
	pbssdk "github.com/certimate-go/certimate/pkg/sdk3rd/proxmoxbs"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// Proxmox BS 服务地址。
	ServerUrl string `json:"serverUrl"`
	// Proxmox BS API Token。
	ApiToken string `json:"apiToken"`
	// Proxmox BS API Token Secret。
	ApiTokenSecret string `json:"apiTokenSecret"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 节点名称。
	NodeName string `json:"nodeName"`
	// 是否自动重启。
	AutoRestart bool `json:"autoRestart"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *pbssdk.Client
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.ApiToken, config.ApiTokenSecret, config.AllowInsecureConnections)
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
	if d.config.NodeName == "" {
		return nil, fmt.Errorf("config `nodeName` is required")
	}

	// 上传自定义证书
	// REF: https://pbs.proxmox.com/docs/api-viewer/index.html#/nodes/{node}/certificates/custom
	nodeUploadCustomCertificateReq := &pbssdk.NodeUploadCustomCertificateRequest{
		Certificates: certPEM,
		Key:          privkeyPEM,
		Force:        true,
		Restart:      d.config.AutoRestart,
	}
	nodeUploadCustomCertificateResp, err := d.sdkClient.NodeUploadCustomCertificateWithContext(ctx, d.config.NodeName, nodeUploadCustomCertificateReq)
	d.logger.Debug("sdk request 'node.UploadCustomCertificate'", slog.String("params.node", d.config.NodeName), slog.Any("request", nodeUploadCustomCertificateReq), slog.Any("response", nodeUploadCustomCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'node.UploadCustomCertificate': %w", err)
	}

	return &DeployResult{}, nil
}

func createSDKClient(serverUrl, apiToken, apiTokenSecret string, skipTlsVerify bool) (*pbssdk.Client, error) {
	client, err := pbssdk.NewClient(
		serverUrl,
		pbssdk.WithApiToken(apiToken, apiTokenSecret),
	)
	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, nil
}
