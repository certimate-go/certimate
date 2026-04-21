package tencentcloudclb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/samber/lo"
	tcclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/tencentcloud-ssl"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-clb/internal"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type DeployerConfig struct {
	// 腾讯云 SecretId。
	SecretId string `json:"secretId"`
	// 腾讯云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 腾讯云接口端点。
	Endpoint string `json:"endpoint,omitempty"`
	// 腾讯云地域。
	Region string `json:"region"`
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// 负载均衡器 ID。
	// 部署资源类型为 [RESOURCE_TYPE_SSLDEPLOY]、[RESOURCE_TYPE_LOADBALANCER]、[RESOURCE_TYPE_RULEDOMAIN] 时必填。
	LoadbalancerId string `json:"loadbalancerId,omitempty"`
	// 负载均衡监听 ID。
	// 部署资源类型为 [RESOURCE_TYPE_SSLDEPLOY]、[RESOURCE_TYPE_LOADBALANCER]、[RESOURCE_TYPE_LISTENER]、[RESOURCE_TYPE_RULEDOMAIN] 时必填。
	ListenerId string `json:"listenerId,omitempty"`
	// SNI 域名或七层转发规则域名（支持泛域名）。
	// 部署资源类型为 [RESOURCE_TYPE_SSLDEPLOY] 时选填；部署资源类型为 [RESOURCE_TYPE_RULEDOMAIN] 时必填。
	Domain string `json:"domain,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *internal.ClbClient
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.SecretId, config.SecretKey, config.Endpoint, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		SecretId:  config.SecretId,
		SecretKey: config.SecretKey,
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM, issuerCertificatePEM string) (*deployer.DeployResult, error) {
	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 根据部署资源类型决定部署方式
	switch d.config.ResourceType {
	case RESOURCE_TYPE_LOADBALANCER:
		if err := d.deployToLoadbalancer(ctx, upres.CertId); err != nil {
			return nil, err
		}

	case RESOURCE_TYPE_LISTENER:
		if err := d.deployToListener(ctx, upres.CertId); err != nil {
			return nil, err
		}

	case RESOURCE_TYPE_RULEDOMAIN:
		if err := d.deployToRuleDomain(ctx, upres.CertId); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported resource type '%s'", d.config.ResourceType)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) deployToLoadbalancer(ctx context.Context, cloudCertId string) error {
	if d.config.LoadbalancerId == "" {
		return errors.New("config `loadbalancerId` is required")
	}

	// 查询监听器列表
	// REF: https://cloud.tencent.com/document/api/214/30686
	listenerIds := make([]string, 0)
	describeListenersReq := tcclb.NewDescribeListenersRequest()
	describeListenersReq.LoadBalancerId = common.StringPtr(d.config.LoadbalancerId)
	describeListenersResp, err := d.sdkClient.DescribeListeners(describeListenersReq)
	d.logger.Debug("sdk request 'clb.DescribeListeners'", slog.Any("request", describeListenersReq), slog.Any("response", describeListenersResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'clb.DescribeListeners': %w", err)
	} else {
		if describeListenersResp.Response.Listeners != nil {
			for _, listener := range describeListenersResp.Response.Listeners {
				if listener.Protocol == nil || (*listener.Protocol != "HTTPS" && *listener.Protocol != "TCP_SSL" && *listener.Protocol != "QUIC") {
					continue
				}

				listenerIds = append(listenerIds, *listener.ListenerId)
			}
		}
	}

	// 遍历更新监听器证书
	if len(listenerIds) == 0 {
		d.logger.Info("no clb listeners to deploy")
	} else {
		d.logger.Info("found https/tcpssl/quic listeners to deploy", slog.Any("listenerIds", listenerIds))
		var errs []error

		for _, listenerId := range listenerIds {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := d.updateListenerCertificate(ctx, d.config.LoadbalancerId, listenerId, cloudCertId); err != nil {
					errs = append(errs, err)
				}
			}
		}

		if len(errs) > 0 {
			return errors.Join(errs...)
		}
	}

	return nil
}

func (d *Deployer) deployToListener(ctx context.Context, cloudCertId string) error {
	if d.config.LoadbalancerId == "" {
		return errors.New("config `loadbalancerId` is required")
	}
	if d.config.ListenerId == "" {
		return errors.New("config `listenerId` is required")
	}

	// 更新监听器证书
	if err := d.updateListenerCertificate(ctx, d.config.LoadbalancerId, d.config.ListenerId, cloudCertId); err != nil {
		return err
	}

	return nil
}

func (d *Deployer) deployToRuleDomain(ctx context.Context, cloudCertId string) error {
	if d.config.LoadbalancerId == "" {
		return errors.New("config `loadbalancerId` is required")
	}
	if d.config.ListenerId == "" {
		return errors.New("config `listenerId` is required")
	}
	if d.config.Domain == "" {
		return errors.New("config `domain` is required")
	}

	// 修改负载均衡七层监听器转发规则的域名级别属性
	// REF: https://cloud.tencent.com/document/api/214/38092
	modifyDomainAttributesReq := tcclb.NewModifyDomainAttributesRequest()
	modifyDomainAttributesReq.LoadBalancerId = common.StringPtr(d.config.LoadbalancerId)
	modifyDomainAttributesReq.ListenerId = common.StringPtr(d.config.ListenerId)
	modifyDomainAttributesReq.Domain = common.StringPtr(d.config.Domain)
	modifyDomainAttributesReq.Certificate = &tcclb.CertificateInput{
		SSLMode: common.StringPtr("UNIDIRECTIONAL"),
		CertId:  common.StringPtr(cloudCertId),
	}
	modifyDomainAttributesResp, err := d.sdkClient.ModifyDomainAttributes(modifyDomainAttributesReq)
	d.logger.Debug("sdk request 'clb.ModifyDomainAttributes'", slog.Any("request", modifyDomainAttributesReq), slog.Any("response", modifyDomainAttributesResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'clb.ModifyDomainAttributes': %w", err)
	}

	// 查询异步任务状态，等待任务状态变更
	// REF: https://cloud.tencent.com/document/product/214/30683
	if _, err := xwait.UntilWithContext(ctx, func(_ context.Context, _ int) (bool, error) {
		describeTaskStatusReq := tcclb.NewDescribeTaskStatusRequest()
		describeTaskStatusReq.TaskId = modifyDomainAttributesResp.Response.RequestId
		describeTaskStatusResp, err := d.sdkClient.DescribeTaskStatus(describeTaskStatusReq)
		d.logger.Debug("sdk request 'clb.DescribeTaskStatus'", slog.Any("request", describeTaskStatusReq), slog.Any("response", describeTaskStatusResp))
		if err != nil {
			return false, fmt.Errorf("failed to execute sdk request 'clb.DescribeTaskStatus': %w", err)
		}

		switch lo.FromPtr(describeTaskStatusResp.Response.Status) {
		case 0:
			return true, nil
		case 1:
			return false, fmt.Errorf("unexpected tencentcloud task status")
		}

		d.logger.Info("waiting for tencentcloud task completion ...")
		return false, nil
	}, time.Second*5); err != nil {
		return err
	}

	return nil
}

func (d *Deployer) updateListenerCertificate(ctx context.Context, cloudLoadbalancerId, cloudListenerId, cloudCertId string) error {
	// 查询负载均衡的监听器列表
	// REF: https://cloud.tencent.com/document/api/214/30686
	describeListenersReq := tcclb.NewDescribeListenersRequest()
	describeListenersReq.LoadBalancerId = common.StringPtr(cloudLoadbalancerId)
	describeListenersReq.ListenerIds = common.StringPtrs([]string{cloudListenerId})
	describeListenersResp, err := d.sdkClient.DescribeListeners(describeListenersReq)
	d.logger.Debug("sdk request 'clb.DescribeListeners'", slog.Any("request", describeListenersReq), slog.Any("response", describeListenersResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'clb.DescribeListeners': %w", err)
	} else if len(describeListenersResp.Response.Listeners) == 0 {
		return fmt.Errorf("could not find listener '%s'", cloudListenerId)
	}

	// 修改监听器属性
	// REF: https://cloud.tencent.com/document/api/214/30681
	modifyListenerReq := tcclb.NewModifyListenerRequest()
	modifyListenerReq.LoadBalancerId = common.StringPtr(cloudLoadbalancerId)
	modifyListenerReq.ListenerId = common.StringPtr(cloudListenerId)
	modifyListenerReq.Certificate = &tcclb.CertificateInput{CertId: common.StringPtr(cloudCertId)}
	if describeListenersResp.Response.Listeners[0].Certificate != nil && describeListenersResp.Response.Listeners[0].Certificate.SSLMode != nil {
		modifyListenerReq.Certificate.SSLMode = describeListenersResp.Response.Listeners[0].Certificate.SSLMode
		modifyListenerReq.Certificate.CertCaId = describeListenersResp.Response.Listeners[0].Certificate.CertCaId
	} else {
		modifyListenerReq.Certificate.SSLMode = common.StringPtr("UNIDIRECTIONAL")
	}
	modifyListenerResp, err := d.sdkClient.ModifyListener(modifyListenerReq)
	d.logger.Debug("sdk request 'clb.ModifyListener'", slog.Any("request", modifyListenerReq), slog.Any("response", modifyListenerResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'clb.ModifyListener': %w", err)
	}

	// 查询异步任务状态，等待任务状态变更
	// REF: https://cloud.tencent.com/document/product/214/30683
	if _, err := xwait.UntilWithContext(ctx, func(_ context.Context, _ int) (bool, error) {
		describeTaskStatusReq := tcclb.NewDescribeTaskStatusRequest()
		describeTaskStatusReq.TaskId = modifyListenerResp.Response.RequestId
		describeTaskStatusResp, err := d.sdkClient.DescribeTaskStatus(describeTaskStatusReq)
		d.logger.Debug("sdk request 'clb.DescribeTaskStatus'", slog.Any("request", describeTaskStatusReq), slog.Any("response", describeTaskStatusResp))
		if err != nil {
			return false, fmt.Errorf("failed to execute sdk request 'clb.DescribeTaskStatus': %w", err)
		}

		switch lo.FromPtr(describeTaskStatusResp.Response.Status) {
		case 0:
			return true, nil
		case 1:
			return false, fmt.Errorf("unexpected tencentcloud task status")
		}

		d.logger.Info("waiting for tencentcloud task completion ...")
		return false, nil
	}, time.Second*5); err != nil {
		return err
	}

	return nil
}

func createSDKClient(secretId, secretKey, endpoint, region string) (*internal.ClbClient, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := internal.NewClbClient(credential, region, cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
