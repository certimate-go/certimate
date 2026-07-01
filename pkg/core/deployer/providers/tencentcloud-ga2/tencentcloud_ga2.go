package tencentcloudga2

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcerrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	tcga2 "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ga2/v20250115"
	tcssl "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/tencentcloud-ssl"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
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
	// 全球加速实例 ID。
	// 部署目标为 [DEPLOY_TARGET_LISTENER] 时必填。
	AcceleratorId string `json:"acceleratorId,omitempty"`
	// 监听器 ID。
	// 部署目标为 [DEPLOY_TARGET_LISTENER] 时必填。
	ListenerId string `json:"listenerId,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClients *wSDKClients
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

type wSDKClients struct {
	GA2 *tcga2.Client
	SSL *tcssl.Client
}

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	clients, err := createSDKClients(config.SecretId, config.SecretKey, config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
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
		sdkClients: clients,
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*DeployResult, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

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
		if err := d.deployToListener(ctx, upres.CertId, certX509.DNSNames); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported deploy target '%s'", d.config.DeployTarget)
	}

	return &DeployResult{}, nil
}

func (d *Deployer) deployToListener(ctx context.Context, cloudCertId string, cloudCertSANs []string) error {
	if d.config.AcceleratorId == "" {
		return fmt.Errorf("config `acceleratorId` is required")
	}
	if d.config.ListenerId == "" {
		return fmt.Errorf("config `listenerId` is required")
	}

	// 更新监听器证书
	if err := d.updateListenerCertificate(ctx, d.config.AcceleratorId, d.config.ListenerId, cloudCertId, cloudCertSANs); err != nil {
		return err
	}

	return nil
}

func (d *Deployer) updateListenerCertificate(ctx context.Context, cloudAcceleratorId, cloudListenerId, cloudCertId string, cloudCertSANs []string) error {
	// 查询监听器
	// REF: https://cloud.tencent.com/document/product/1817/130160
	describeListenersReq := tcga2.NewDescribeListenersRequest()
	describeListenersReq.GlobalAcceleratorId = common.StringPtr(cloudAcceleratorId)
	describeListenersReq.Filters = []*tcga2.Filter{
		{
			Name:   common.StringPtr("listener-id"),
			Values: []*string{common.StringPtr(cloudListenerId)},
		},
	}
	describeListenersReq.Offset = common.Uint64Ptr(0)
	describeListenersReq.Limit = common.Uint64Ptr(1)
	describeListenersResp, err := d.sdkClients.GA2.DescribeListenersWithContext(ctx, describeListenersReq)
	d.logger.Debug("sdk request 'ga2.DescribeListeners'", slog.Any("request", describeListenersReq), slog.Any("response", describeListenersResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'ga2.DescribeListeners': %w", err)
	} else if len(describeListenersResp.Response.ListenerSet) == 0 {
		return fmt.Errorf("could not find listener '%s'", cloudListenerId)
	}

	// 获取证书信息，避免重复绑定
	// REF: https://cloud.tencent.com/document/api/400/41674
	serverCertificateIds := make([]string, 0)
	for _, serverCertificateId := range describeListenersResp.Response.ListenerSet[0].ServerCertificates {
		if cloudCertId == lo.FromPtr(serverCertificateId) {
			return nil
		}

		describeCertificateReq := tcssl.NewDescribeCertificateRequest()
		describeCertificateReq.CertificateId = serverCertificateId
		describeCertificateResp, err := d.sdkClients.SSL.DescribeCertificateWithContext(ctx, describeCertificateReq)
		d.logger.Debug("sdk request 'ssl.DescribeCertificate'", slog.Any("request", describeCertificateReq), slog.Any("response", describeCertificateResp))
		if err != nil {
			if sdkErr, ok := err.(*tcerrors.TencentCloudSDKError); ok {
				if sdkErrCode := sdkErr.Code; sdkErrCode == "FailedOperation.CertificateNotFound" {
					continue
				}
			}

			return fmt.Errorf("failed to execute sdk request 'ssl.DescribeCertificate': %w", err)
		} else {
			certSANMatched := lo.ElementsMatch(lo.FromSlicePtr(describeCertificateResp.Response.SubjectAltName), cloudCertSANs)
			if certSANMatched { // 同域名证书需要删除
				continue
			}

			serverCertificateIds = append(serverCertificateIds, lo.FromPtr(serverCertificateId))
		}
	}

	// 修改监听器
	// REF: https://cloud.tencent.com/document/product/1817/130155
	modifyListenerReq := tcga2.NewModifyListenerRequest()
	modifyListenerReq.GlobalAcceleratorId = common.StringPtr(cloudAcceleratorId)
	modifyListenerReq.ListenerId = common.StringPtr(cloudListenerId)
	modifyListenerReq.ServerCertificates = common.StringPtrs(append(serverCertificateIds, cloudCertId))
	modifyListenerResp, err := d.sdkClients.GA2.ModifyListenerWithContext(ctx, modifyListenerReq)
	d.logger.Debug("sdk request 'ga2.ModifyListener'", slog.Any("request", modifyListenerReq), slog.Any("response", modifyListenerResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'ga2.ModifyListener': %w", err)
	}

	return nil
}

func createSDKClients(secretId, secretKey, endpoint string) (*wSDKClients, error) {
	wsdk := &wSDKClients{}

	{
		credential := common.NewCredential(secretId, secretKey)

		cpf := profile.NewClientProfile()
		if endpoint != "" {
			cpf.HttpProfile.Endpoint = endpoint
		}

		client, err := tcga2.NewClient(credential, "", cpf)
		if err != nil {
			return nil, err
		}

		wsdk.GA2 = client
	}

	{
		credential := common.NewCredential(secretId, secretKey)

		cpf := profile.NewClientProfile()
		if strings.HasSuffix(endpoint, "intl.tencentcloudapi.com") {
			cpf.HttpProfile.Endpoint = "ssl.intl.tencentcloudapi.com"
		}

		client, err := tcssl.NewClient(credential, "", cpf)
		if err != nil {
			return nil, err
		}

		wsdk.SSL = client
	}

	return wsdk, nil
}
