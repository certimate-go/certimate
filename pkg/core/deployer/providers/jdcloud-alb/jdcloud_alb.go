package jdcloudalb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	jdcore "github.com/jdcloud-api/jdcloud-sdk-go/core"
	jdcommon "github.com/jdcloud-api/jdcloud-sdk-go/services/common/models"
	jdlb "github.com/jdcloud-api/jdcloud-sdk-go/services/lb/apis"
	jdlbmodel "github.com/jdcloud-api/jdcloud-sdk-go/services/lb/models"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/jdcloud-ssl"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/jdcloud-alb/internal"
)

type DeployerConfig struct {
	// 京东云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 京东云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 京东云地域 ID。
	RegionId string `json:"regionId"`
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// 负载均衡器 ID。
	// 部署资源类型为 [RESOURCE_TYPE_LOADBALANCER] 时必填。
	LoadbalancerId string `json:"loadbalancerId,omitempty"`
	// 监听器 ID。
	// 部署资源类型为 [RESOURCE_TYPE_LISTENER] 时必填。
	ListenerId string `json:"listenerId,omitempty"`
	// SNI 域名（支持泛域名）。
	// 部署资源类型为 [RESOURCE_TYPE_LOADBALANCER]、[RESOURCE_TYPE_LISTENER] 时选填。
	Domain string `json:"domain,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *internal.LbClient
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
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

	default:
		return nil, fmt.Errorf("unsupported resource type '%s'", d.config.ResourceType)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) deployToLoadbalancer(ctx context.Context, cloudCertId string) error {
	if d.config.LoadbalancerId == "" {
		return errors.New("config `loadbalancerId` is required")
	}

	// 查询负载均衡器详情
	// REF: https://docs.jdcloud.com/cn/load-balancer/api/describeloadbalancer
	describeLoadBalancerReq := jdlb.NewDescribeLoadBalancerRequestWithoutParam()
	describeLoadBalancerReq.SetRegionId(d.config.RegionId)
	describeLoadBalancerReq.SetLoadBalancerId(d.config.LoadbalancerId)
	describeLoadBalancerResp, err := d.sdkClient.DescribeLoadBalancer(describeLoadBalancerReq)
	d.logger.Debug("sdk request 'lb.DescribeLoadBalancer'", slog.Any("request", describeLoadBalancerReq), slog.Any("response", describeLoadBalancerResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'lb.DescribeLoadBalancer': %w", err)
	}

	// 查询监听器列表
	// REF: https://docs.jdcloud.com/cn/load-balancer/api/describelisteners
	listenerIds := make([]string, 0)
	describeListenersPageNumber := 1
	describeListenersPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		describeListenersReq := jdlb.NewDescribeListenersRequestWithoutParam()
		describeListenersReq.SetRegionId(d.config.RegionId)
		describeListenersReq.SetFilters([]jdcommon.Filter{{Name: "loadBalancerId", Values: []string{d.config.LoadbalancerId}}})
		describeListenersReq.SetPageSize(describeListenersPageNumber)
		describeListenersReq.SetPageSize(describeListenersPageSize)
		describeListenersResp, err := d.sdkClient.DescribeListeners(describeListenersReq)
		d.logger.Debug("sdk request 'lb.DescribeListeners'", slog.Any("request", describeListenersReq), slog.Any("response", describeListenersResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'lb.DescribeListeners': %w", err)
		}

		for _, listener := range describeListenersResp.Result.Listeners {
			if strings.EqualFold(listener.Protocol, "https") || strings.EqualFold(listener.Protocol, "tls") {
				listenerIds = append(listenerIds, listener.ListenerId)
			}
		}

		if len(describeListenersResp.Result.Listeners) < describeListenersPageSize {
			break
		}

		describeListenersPageNumber++
	}

	// 遍历更新监听器证书
	if len(listenerIds) == 0 {
		d.logger.Info("no listeners to deploy")
	} else {
		d.logger.Info("found https/tls listeners to deploy", slog.Any("listenerIds", listenerIds))

		var errs []error

		for _, listenerId := range listenerIds {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := d.updateListenerCertificate(ctx, listenerId, cloudCertId); err != nil {
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
	if d.config.ListenerId == "" {
		return errors.New("config `listenerId` is required")
	}

	// 更新监听器证书
	if err := d.updateListenerCertificate(ctx, d.config.ListenerId, cloudCertId); err != nil {
		return err
	}

	return nil
}

func (d *Deployer) updateListenerCertificate(ctx context.Context, cloudListenerId string, cloudCertId string) error {
	// 查询监听器详情
	// REF: https://docs.jdcloud.com/cn/load-balancer/api/describelistener
	describeListenerReq := jdlb.NewDescribeListenerRequestWithoutParam()
	describeListenerReq.SetRegionId(d.config.RegionId)
	describeListenerReq.SetListenerId(cloudListenerId)
	describeListenerResp, err := d.sdkClient.DescribeListener(describeListenerReq)
	d.logger.Debug("sdk request 'lb.DescribeListener'", slog.Any("request", describeListenerReq), slog.Any("response", describeListenerResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'lb.DescribeListener': %w", err)
	}

	if d.config.Domain == "" {
		// 未指定 SNI，只需部署到监听器

		// 修改监听器信息
		// REF: https://docs.jdcloud.com/cn/load-balancer/api/updatelistener
		updateListenerReq := jdlb.NewUpdateListenerRequestWithoutParam()
		updateListenerReq.SetRegionId(d.config.RegionId)
		updateListenerReq.SetListenerId(cloudListenerId)
		updateListenerReq.SetCertificateSpecs([]jdlbmodel.CertificateSpec{{CertificateId: cloudCertId}})
		updateListenerResp, err := d.sdkClient.UpdateListener(updateListenerReq)
		d.logger.Debug("sdk request 'lb.UpdateListener'", slog.Any("request", updateListenerReq), slog.Any("response", updateListenerResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'lb.UpdateListener': %w", err)
		}
	} else {
		// 指定 SNI，需部署到扩展证书

		extCertSpecs := lo.Filter(describeListenerResp.Result.Listener.ExtensionCertificateSpecs, func(extCertSpec jdlbmodel.ExtensionCertificateSpec, _ int) bool {
			return extCertSpec.Domain == d.config.Domain
		})
		if len(extCertSpecs) == 0 {
			return errors.New("could not find any extension certificates")
		}

		// 批量修改扩展证书
		// REF: https://docs.jdcloud.com/cn/load-balancer/api/updatelistenercertificates
		updateListenerCertificatesReq := jdlb.NewUpdateListenerCertificatesRequestWithoutParam()
		updateListenerCertificatesReq.SetRegionId(d.config.RegionId)
		updateListenerCertificatesReq.SetListenerId(cloudListenerId)
		updateListenerCertificatesReq.SetCertificates(lo.Map(extCertSpecs, func(extCertSpec jdlbmodel.ExtensionCertificateSpec, _ int) jdlbmodel.ExtCertificateUpdateSpec {
			return jdlbmodel.ExtCertificateUpdateSpec{
				CertificateBindId: extCertSpec.CertificateBindId,
				CertificateId:     &cloudCertId,
				Domain:            &extCertSpec.Domain,
			}
		}))
		updateListenerCertificatesResp, err := d.sdkClient.UpdateListenerCertificates(updateListenerCertificatesReq)
		d.logger.Debug("sdk request 'lb.UpdateListenerCertificates'", slog.Any("request", updateListenerCertificatesReq), slog.Any("response", updateListenerCertificatesResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'lb.UpdateListenerCertificates': %w", err)
		}
	}

	return nil
}

func createSDKClient(accessKeyId, accessKeySecret string) (*internal.LbClient, error) {
	clientCredentials := jdcore.NewCredentials(accessKeyId, accessKeySecret)
	client := internal.NewLbClient(clientCredentials)
	return client, nil
}
