package qingcloudlb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/samber/lo"
	qcconfig "github.com/yunify/qingcloud-sdk-go/config"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/qingcloud-lb"
	qclbsdk "github.com/certimate-go/certimate/pkg/sdk3rd/qingcloud/lb"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// 青云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 青云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 青云区域 ID。
	ZoneId string `json:"zoneId"`
	// 部署目标。
	DeployTarget string `json:"deployTarget"`
	// 负载均衡器 ID。
	// 部署目标为 [DEPLOY_TARGET_LOADBALANCER] 时必填。
	LoadbalancerId string `json:"loadbalancerId,omitempty"`
	// 负载均衡监听 ID。
	// 部署目标为 [DEPLOY_TARGET_LISTENER] 时必填。
	ListenerId string `json:"listenerId,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *qclbsdk.LoadBalancerService
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey, config.ZoneId)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		SecretAccessKey: config.SecretAccessKey,
		ZoneId:          config.ZoneId,
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*DeployResult, error) {
	// 根据部署目标决定业务流程
	switch d.config.DeployTarget {
	case DEPLOY_TARGET_LOADBALANCER:
		if err := d.deployToLoadbalancer(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	case DEPLOY_TARGET_LISTENER:
		if err := d.deployToListener(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported deploy target '%s'", d.config.DeployTarget)
	}

	return &DeployResult{}, nil
}

func (d *Deployer) deployToLoadbalancer(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.LoadbalancerId == "" {
		return fmt.Errorf("config `loadbalancerId` is required")
	}

	// 获取负载均衡器的监听器列表
	// REF: https://docsv4.qingcloud.com/user_guide/development_docs/api/api_list/network/loadbalancer/describe_loadbalancer_listeners/
	listenerIds := make([]string, 0)
	describeLoadBalancerListenersOffset := 0
	describeLoadBalancerListenersLimit := 100
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		listListenersReq := &qclbsdk.DescribeLoadBalancerListenersInput{
			LoadBalancer: lo.ToPtr(d.config.LoadbalancerId),
			Offset:       lo.ToPtr(describeLoadBalancerListenersOffset),
			Limit:        lo.ToPtr(describeLoadBalancerListenersLimit),
		}
		listListenersResp, err := d.sdkClient.DescribeLoadBalancerListeners(listListenersReq)
		d.logger.Debug("sdk request 'lb.DescribeLoadBalancerListeners'", slog.Any("request", listListenersReq), slog.Any("response", listListenersResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'lb.DescribeLoadBalancerListeners': %w", err)
		}

		for _, listener := range listListenersResp.LoadBalancerListenerSet {
			if !strings.EqualFold(lo.FromPtr(listener.ListenerProtocol), "https") {
				continue
			}

			listenerIds = append(listenerIds, lo.FromPtr(listener.LoadBalancerListenerID))
		}

		if len(listListenersResp.LoadBalancerListenerSet) < describeLoadBalancerListenersLimit {
			break
		}

		describeLoadBalancerListenersOffset += describeLoadBalancerListenersLimit
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 遍历更新监听器证书
	if len(listenerIds) == 0 {
		d.logger.Info("no listeners to deploy")
	} else {
		d.logger.Info("found https listeners to deploy", slog.Any("listenerIds", listenerIds))
		var errs []error

		for _, listenerId := range listenerIds {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := d.updateListenerCertificate(ctx, listenerId, upres.CertId); err != nil {
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

func (d *Deployer) deployToListener(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.ListenerId == "" {
		return fmt.Errorf("config `listenerId` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 更新监听器证书
	if err := d.updateListenerCertificate(ctx, d.config.ListenerId, upres.CertId); err != nil {
		return err
	}

	return nil
}

func (d *Deployer) updateListenerCertificate(ctx context.Context, cloudListenerId string, cloudCertId string) error {
	// 绑定服务器证书到负载均衡监听器上
	// REF: https://docsv4.qingcloud.com/user_guide/development_docs/api/api_list/network/loadbalancer/bind_certs_to_listener/
	associateServerCertsToLBListenerReq := &qclbsdk.AssociateServerCertsToLBListenerInput{
		LoadBalancerListener: lo.ToPtr(cloudListenerId),
		ServerCertificates:   []*string{lo.ToPtr(cloudCertId)},
	}
	associateServerCertsToLBListenerResp, err := d.sdkClient.AssociateServerCertsToLBListener(associateServerCertsToLBListenerReq)
	d.logger.Debug("sdk request 'lb.AssociateServerCertsToLBListener'", slog.Any("request", associateServerCertsToLBListenerReq), slog.Any("response", associateServerCertsToLBListenerResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'lb.AssociateServerCertsToLBListener': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, secretAccessKey, zoneId string) (*qclbsdk.LoadBalancerService, error) {
	config, err := qcconfig.New(accessKeyId, secretAccessKey)
	if err != nil {
		return nil, err
	} else {
		config.Zone = zoneId
	}

	service, err := qclbsdk.NewService(config)
	if err != nil {
		return nil, err
	}

	return service, nil
}
