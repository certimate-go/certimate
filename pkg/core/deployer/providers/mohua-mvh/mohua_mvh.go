package mohuamvh

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	mohuasdk "github.com/mohuatech/mohuacloud-go-sdk"
	"github.com/mohuatech/mohuacloud-go-sdk/types"
)

// DeployerConfig 嘿华云虚拟主机部署器配置
type DeployerConfig struct {
	// 嘿华云账号
	AccessKey string `json:"accessKey"`
	// 嘿华云API密钥
	SecretKey string `json:"secretKey"`
	// 虚拟主机ID
	HostID string `json:"hostID"`
	// 域名ID
	DomainID string `json:"domainID"`
}

// Deployer 嘿华云虚拟主机部署器
type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *mohuasdk.Client
}

var _ deployer.Provider = (*Deployer)(nil)

// NewDeployer 创建嘿华云部署器
func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	// 验证必要参数
	if config.AccessKey == "" {
		return nil, errors.New("config `accessKey` is required")
	}
	if config.SecretKey == "" {
		return nil, errors.New("config `secretKey` is required")
	}
	if config.HostID == "" {
		return nil, errors.New("config `hostID` is required")
	}
	if config.DomainID == "" {
		return nil, errors.New("config `domainID` is required")
	}

	// 创建SDK客户端
	client := mohuasdk.NewClient(
		mohuasdk.WithCredentials(config.AccessKey, config.SecretKey),
	)

	return &Deployer{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

// SetLogger 设置日志记录器
func (d *Deployer) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}
}

// Deploy 部署证书到嘿华云虚拟主机
func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	// 1. 首先登录获取token
	_, err := d.sdkClient.Auth.Login("", "")
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}
	
	// 2. 设置SSL证书
	err = d.setSSL(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("Failed to set SSL certificate: %w", err)
	}

	return &deployer.DeployResult{}, nil
}

// setSSL 设置SSL证书到指定域名
func (d *Deployer) setSSL(ctx context.Context, certPEM, privkeyPEM string) error {
	// 根据官方demo，domainID 应该是整数类型
	// 我们需要将 string 类型的 DomainID 转换为 int
	domainIDInt, err := parseDomainID(d.config.DomainID)
	if err != nil {
		return fmt.Errorf("Invalid domain ID format: %w", err)
	}

	// 构建SSL设置请求
	sslRequest := &types.SetSSLRequest{
		ID:      domainIDInt, // 使用转换后的整数ID
		SSLCert: certPEM,
		SSLKey:  privkeyPEM,
	}

	// 调用SDK设置SSL
	_, err = d.sdkClient.VirtualHost.SetSSL(d.config.HostID, sslRequest)
	if err != nil {
		return fmt.Errorf("Failed to call SDK to set SSL: %w", err)
	}

	return nil
}

// parseDomainID 将字符串类型的域名ID转换为整数
func parseDomainID(domainID string) (int, error) {
	var id int
	_, err := fmt.Sscanf(domainID, "%d", &id)
	if err != nil {
		return 0, fmt.Errorf("Domain ID '%s' is not a valid integer: %w", domainID, err)
	}
	return id, nil
}