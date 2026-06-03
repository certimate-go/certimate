package tencentcloudgaap

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	certmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/tencentcloud-ssl"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	tcgaap "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/gaap/v20180529"
)

type DeployerConfig struct {
	// 腾讯云 SecretId。
	SecretId string `json:"secretId"`
	// 腾讯云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 腾讯云项目 ID。
	ProjectId int64 `json:"projectId,omitempty"`
	// 腾讯云接口端点。
	Endpoint string `json:"endpoint,omitempty"`
	// 部署目标。
	DeployTarget string `json:"deployTarget"`
	// 通道 ID。
	// 选填。
	ProxyId string `json:"proxyId,omitempty"`
	// 负载均衡监听 ID。
	// 部署目标为 [DEPLOY_TARGET_LISTENER] 时必填。
	ListenerId string `json:"listenerId,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *tcgaap.Client
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClients(config.SecretId, config.SecretKey, config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := certmgrimpl.NewCertmgr(&certmgrimpl.CertmgrConfig{
		SecretId:  config.SecretId,
		SecretKey: config.SecretKey,
		ProjectId: config.ProjectId,
		Endpoint: lo.
			If(strings.HasSuffix(config.Endpoint, "intl.tencentcloudapi.com"), "ssl.intl.tencentcloudapi.com"). // 国际站使用独立的接口端点
			Else(""),
	})
	if err != nil {
		return nil, fmt.Errorf("could not create certmgr: %w", err)
	}

	return &Deployer{
		config:     config,
		logger:     slog.Default(),
		sdkClient:  client,
		sdkCertmgr: pcertmgr,
	}, nil
}

func (d *Deployer) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}

	d.sdkCertmgr.SetLogger(logger)
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 根据部署目标决定业务流程
	switch d.config.DeployTarget {
	case DEPLOY_TARGET_LISTENER:
		if err := d.deployToListener(ctx, upres.CertId); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported deploy target '%s'", d.config.DeployTarget)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) deployToListener(ctx context.Context, cloudCertId string) error {
	if d.config.ListenerId == "" {
		return fmt.Errorf("config `listenerId` is required")
	}

	// 更新监听器证书
	if err := d.updateHttpsListenerCertificate(ctx, d.config.ListenerId, cloudCertId); err != nil {
		return err
	}

	return nil
}

func (d *Deployer) updateHttpsListenerCertificate(ctx context.Context, cloudListenerId, cloudCertId string) error {
	// 查询 HTTPS 监听器信息
	// REF: https://cloud.tencent.com/document/api/608/37001
	describeHTTPSListenersReq := tcgaap.NewDescribeHTTPSListenersRequest()
	describeHTTPSListenersReq.ListenerId = common.StringPtr(cloudListenerId)
	describeHTTPSListenersReq.Offset = common.Uint64Ptr(0)
	describeHTTPSListenersReq.Limit = common.Uint64Ptr(1)
	describeHTTPSListenersResp, err := d.sdkClient.DescribeHTTPSListenersWithContext(ctx, describeHTTPSListenersReq)
	d.logger.Debug("sdk request 'gaap.DescribeHTTPSListeners'", slog.Any("request", describeHTTPSListenersReq), slog.Any("response", describeHTTPSListenersResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'gaap.DescribeHTTPSListeners': %w", err)
	} else if len(describeHTTPSListenersResp.Response.ListenerSet) == 0 {
		return fmt.Errorf("could not find listener '%s'", cloudListenerId)
	}

	// 修改 HTTPS 监听器配置
	// REF: https://cloud.tencent.com/document/api/608/36996
	modifyHTTPSListenerAttributeReq := tcgaap.NewModifyHTTPSListenerAttributeRequest()
	modifyHTTPSListenerAttributeReq.ProxyId = lo.EmptyableToPtr(d.config.ProxyId)
	modifyHTTPSListenerAttributeReq.ListenerId = common.StringPtr(cloudListenerId)
	modifyHTTPSListenerAttributeReq.CertificateId = common.StringPtr(cloudCertId)
	modifyHTTPSListenerAttributeResp, err := d.sdkClient.ModifyHTTPSListenerAttributeWithContext(ctx, modifyHTTPSListenerAttributeReq)
	d.logger.Debug("sdk request 'gaap.ModifyHTTPSListenerAttribute'", slog.Any("request", modifyHTTPSListenerAttributeReq), slog.Any("response", modifyHTTPSListenerAttributeResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'gaap.ModifyHTTPSListenerAttribute': %w", err)
	}

	return nil
}

func createSDKClients(secretId, secretKey, endpoint string) (*tcgaap.Client, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := tcgaap.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
