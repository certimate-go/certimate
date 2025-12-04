package mohuamvh

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	mohuasdk "github.com/mohuatech/mohuacloud-go-sdk"
	mohuasdktypes "github.com/mohuatech/mohuacloud-go-sdk/types"

	"github.com/certimate-go/certimate/pkg/core/deployer"
)

type DeployerConfig struct {
	// 嘿华云账号。
	Username string `json:"username"`
	// 嘿华云 API 密钥。
	ApiPassword string `json:"apiPassword"`
	// 虚拟主机 ID。
	HostId string `json:"hostId"`
	// 域名 ID。
	DomainId string `json:"domainId"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *mohuasdk.Client
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	if d.config.HostId == "" {
		return nil, errors.New("config `hostId` is required")
	}
	if d.config.DomainId == "" {
		return nil, errors.New("config `domainId` is required")
	}

	domainId, err := strconv.ParseInt(d.config.DomainId, 10, 64)
	if err != nil {
		return nil, err
	}

	// 登录获取 Token
	_, err = d.sdkClient.Auth.Login("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to login mohua: %w", err)
	}

	// 设置 SSL 证书
	setSSLReq := &mohuasdktypes.SetSSLRequest{
		ID:      int(domainId),
		SSLCert: certPEM,
		SSLKey:  privkeyPEM,
	}
	setSSLResp, err := d.sdkClient.VirtualHost.SetSSL(d.config.HostId, setSSLReq)
	d.logger.Debug("sdk request 'mvh.SetSSL'", slog.Any("request", setSSLReq), slog.Any("response", setSSLResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'mvh.SetSSL': %w", err)
	}

	return &deployer.DeployResult{}, nil
}

func createSDKClient(username, apiPassword string) (*mohuasdk.Client, error) {
	if username == "" {
		return nil, errors.New("invalid mohua username")
	}
	if apiPassword == "" {
		return nil, errors.New("invalid mohua api password")
	}

	client := mohuasdk.NewClient(
		mohuasdk.WithCredentials(username, apiPassword),
	)
	return client, nil
}
