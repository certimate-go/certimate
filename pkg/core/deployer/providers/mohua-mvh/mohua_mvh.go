package mohuamvh

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/pkg/core"
	mohuasdk "github.com/certimate-go/certimate/pkg/sdk3rd/mohua"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// 嘿华云账号。
	Username string `json:"username"`
	// 嘿华云 API 密钥。
	ApiPassword string `json:"apiPassword"`
	// 虚拟主机 ID。
	HostId string `json:"hostId"`
	// 域名 ID。
	DomainId int64 `json:"domainId"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *mohuasdk.Client
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.Username, config.ApiPassword)
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
	if d.config.HostId == "" {
		return nil, fmt.Errorf("config `hostId` is required")
	}
	if d.config.DomainId == 0 {
		return nil, fmt.Errorf("config `domainId` is required")
	}

	// 设置 SSL 证书
	setSSLReq := &mohuasdk.SetVirtualHostSSLRequest{
		ID:      int(d.config.DomainId),
		SSLCert: certPEM,
		SSLKey:  privkeyPEM,
	}
	setSSLResp, err := d.sdkClient.SetVirtualHostSSL(d.config.HostId, setSSLReq)
	d.logger.Debug("sdk request 'mvh.SetSSL'", slog.Any("request", setSSLReq), slog.Any("response", setSSLResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'mvh.SetSSL': %w", err)
	}

	return &DeployResult{}, nil
}

func createSDKClient(username, apiPassword string) (*mohuasdk.Client, error) {
	client, err := mohuasdk.NewClient(
		mohuasdk.WithLogins(username, apiPassword),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
