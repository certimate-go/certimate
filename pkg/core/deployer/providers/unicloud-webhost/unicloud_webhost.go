package unicloudwebhost

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/certimate-go/certimate/pkg/core"
	unicloudsdk "github.com/certimate-go/certimate/pkg/sdk3rd/dcloud/unicloud"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// uniCloud 控制台账号。
	Username string `json:"username"`
	// uniCloud 控制台密码。
	Password string `json:"password"`
	// 服务空间提供商。
	// 可取值 "aliyun"、"alipay"、"tencent"。
	SpaceProvider string `json:"spaceProvider"`
	// 服务空间 ID。
	SpaceId string `json:"spaceId"`
	// 托管网站域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *unicloudsdk.Client
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.Username, config.Password)
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
	if d.config.SpaceProvider == "" {
		return nil, fmt.Errorf("config `spaceProvider` is required")
	}
	if d.config.SpaceId == "" {
		return nil, fmt.Errorf("config `spaceId` is required")
	}
	if d.config.Domain == "" {
		return nil, fmt.Errorf("config `domain` is required")
	}

	// 变更网站证书
	createDomainWithCertReq := &unicloudsdk.CreateDomainWithCertRequest{
		Provider: d.config.SpaceProvider,
		SpaceId:  d.config.SpaceId,
		Domain:   d.config.Domain,
		Cert:     url.QueryEscape(certPEM),
		Key:      url.QueryEscape(privkeyPEM),
	}
	createDomainWithCertResp, err := d.sdkClient.CreateDomainWithCert(createDomainWithCertReq)
	d.logger.Debug("sdk request 'unicloud.host.CreateDomainWithCert'", slog.Any("request", createDomainWithCertReq), slog.Any("response", createDomainWithCertResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'unicloud.host.CreateDomainWithCert': %w", err)
	}

	return &DeployResult{}, nil
}

func createSDKClient(username, password string) (*unicloudsdk.Client, error) {
	return unicloudsdk.NewClient(username, password)
}
