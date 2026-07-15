package cmcccloudvlb

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/samber/lo"
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/config"
	"gitlab.ecloud.com/ecloud/ecloudsdkvlb/model"

	"github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/gitlab.ecloud.com/ecloud/ecloudsdkvlb"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/cmcccloud-vlb"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// 移动云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 移动云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 移动云资源池 ID。
	PoolId string `json:"poolId"`
	// 部署目标。
	DeployTarget string `json:"deployTarget"`
	// 负载均衡实例 ID。
	// 部署目标为 [DEPLOY_TARGET_LOADBALANCER] 时必填。
	LoadbalancerId string `json:"loadbalancerId,omitempty"`
	// 负载均衡监听器 ID。
	// 部署目标为 [DEPLOY_TARGET_LISTENER] 时必填。
	ListenerId string `json:"listenerId,omitempty"`
	// SNI 域名（支持泛域名）。
	// 部署目标为 [DEPLOY_TARGET_LOADBALANCER]、[DEPLOY_TARGET_LISTENER] 时选填。
	Domain string `json:"domain,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *ecloudsdkvlb.Client
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret, config.PoolId)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
		PoolId:          config.PoolId,
		IsDefault:       config.Domain == "",
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
	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 根据部署目标决定业务流程
	switch d.config.DeployTarget {
	case DEPLOY_TARGET_LOADBALANCER:
		if err := d.deployToLoadbalancer(ctx, upres.CertId); err != nil {
			return nil, err
		}

	case DEPLOY_TARGET_LISTENER:
		if err := d.deployToListener(ctx, upres.CertId); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported deploy target '%s'", d.config.DeployTarget)
	}

	return &DeployResult{}, nil
}

func (d *Deployer) deployToLoadbalancer(ctx context.Context, cloudCertId string) error {
	if d.config.LoadbalancerId == "" {
		return fmt.Errorf("config `loadbalancerId` is required")
	}

	// 查询 HTTPS 监听器列表
	// REF: https://ecloud.10086.cn/op-help-center/doc/article/97026
	listenerIds := make([]string, 0)
	listLoadBalanceHTTPSListenerPage := 1
	listLoadBalanceHTTPSListenerPageSize := 10
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		listLoadBalanceHTTPSListenerReq := &model.ListLoadBalanceHTTPSListenerRequest{
			ListLoadBalanceHTTPSListenerPath: &model.ListLoadBalanceHTTPSListenerPath{
				LoadBalanceId: lo.ToPtr(d.config.LoadbalancerId),
			},
			ListLoadBalanceHTTPSListenerQuery: &model.ListLoadBalanceHTTPSListenerQuery{
				Page:     lo.ToPtr(int32(listLoadBalanceHTTPSListenerPage)),
				PageSize: lo.ToPtr(int32(listLoadBalanceHTTPSListenerPageSize)),
			},
		}
		listLoadBalanceHTTPSListenerResp, err := d.sdkClient.ListLoadBalanceHTTPSListener(listLoadBalanceHTTPSListenerReq)
		d.logger.Debug("sdk request 'vlb.ListLoadBalanceHTTPSListener'", slog.Any("request", listLoadBalanceHTTPSListenerReq), slog.Any("response", listLoadBalanceHTTPSListenerResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'vlb.ListLoadBalanceHTTPSListener': %w", err)
		}

		if listLoadBalanceHTTPSListenerResp.Body == nil || listLoadBalanceHTTPSListenerResp.Body.Content == nil {
			break
		}

		for _, listener := range *listLoadBalanceHTTPSListenerResp.Body.Content {
			listenerIds = append(listenerIds, lo.FromPtr(listener.Id))
		}

		if len(*listLoadBalanceHTTPSListenerResp.Body.Content) < listLoadBalanceHTTPSListenerPageSize {
			break
		}

		listLoadBalanceHTTPSListenerPage++
	}

	// 批量更新监听证书
	if len(listenerIds) == 0 {
		d.logger.Info("no vlb listeners to deploy")
	} else {
		d.logger.Info("found vlb listeners to deploy", slog.Any("listenerIds", listenerIds))

		if err := xloop.ForRangeAllWithContext(ctx, listenerIds, func(ctx context.Context, listenerId string, _ int) error {
			return d.updateListenerCertificate(ctx, listenerId, cloudCertId)
		}); err != nil {
			return err
		}
	}

	return nil
}

func (d *Deployer) deployToListener(ctx context.Context, cloudCertId string) error {
	if d.config.ListenerId == "" {
		return fmt.Errorf("config `listenerId` is required")
	}

	// 更新监听
	if err := d.updateListenerCertificate(ctx, d.config.ListenerId, cloudCertId); err != nil {
		return err
	}

	return nil
}

func (d *Deployer) updateListenerCertificate(ctx context.Context, cloudListenerId string, cloudCertId string) error {
	// 查询 HTTPS 监听器信息
	// REF: https://ecloud.10086.cn/op-help-center/doc/article/97026
	var listenerInfo *model.ListLoadBalanceHTTPSListenerResponseContent
	listLoadBalanceHTTPSListenerPage := 1
	listLoadBalanceHTTPSListenerPageSize := 10
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		listLoadBalanceHTTPSListenerReq := &model.ListLoadBalanceHTTPSListenerRequest{
			ListLoadBalanceHTTPSListenerPath: &model.ListLoadBalanceHTTPSListenerPath{
				LoadBalanceId: lo.ToPtr(d.config.LoadbalancerId),
			},
			ListLoadBalanceHTTPSListenerQuery: &model.ListLoadBalanceHTTPSListenerQuery{
				Page:     lo.ToPtr(int32(listLoadBalanceHTTPSListenerPage)),
				PageSize: lo.ToPtr(int32(listLoadBalanceHTTPSListenerPageSize)),
			},
		}
		listLoadBalanceHTTPSListenerResp, err := d.sdkClient.ListLoadBalanceHTTPSListener(listLoadBalanceHTTPSListenerReq)
		d.logger.Debug("sdk request 'vlb.ListLoadBalanceHTTPSListener'", slog.Any("request", listLoadBalanceHTTPSListenerReq), slog.Any("response", listLoadBalanceHTTPSListenerResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'vlb.ListLoadBalanceHTTPSListener': %w", err)
		}

		if listLoadBalanceHTTPSListenerResp.Body == nil || listLoadBalanceHTTPSListenerResp.Body.Content == nil {
			break
		}

		for _, listener := range *listLoadBalanceHTTPSListenerResp.Body.Content {
			if lo.FromPtr(listener.Id) == cloudListenerId {
				listenerInfo = &listener
				break
			}
		}

		if len(*listLoadBalanceHTTPSListenerResp.Body.Content) < listLoadBalanceHTTPSListenerPageSize {
			break
		}

		listLoadBalanceHTTPSListenerPage++
	}
	if listenerInfo == nil {
		return fmt.Errorf("could not find vlb listener '%s'", cloudListenerId)
	}

	if d.config.Domain == "" {
		// 未指定 SNI，只需部署到默认证书
		if lo.FromPtr(listenerInfo.DefaultTlsContainerId) == cloudCertId {
			d.logger.Info("no need to deploy vlb default certificate")
			return nil
		}
		return d.updateListenerDefaultCertificate(ctx, *listenerInfo, cloudCertId)
	} else {
		// 指定 SNI，需部署到 SNI 证书
		if lo.Contains(listenerInfo.SniContainerIdList, cloudCertId) {
			d.logger.Info("no need to deploy vlb sni certificate")
			return nil
		}
		return d.updateListenerSniCertificate(ctx, *listenerInfo, cloudCertId)
	}
}

func (d *Deployer) updateListenerDefaultCertificate(ctx context.Context, cloudListenerInfo model.ListLoadBalanceHTTPSListenerResponseContent, cloudCertId string) error {
	// 修改 HTTPS 监听器
	// REF: https://ecloud.10086.cn/op-help-center/doc/article/97024
	updateListenerReq := &model.UpdateListenerRequest{
		&model.UpdateListenerBody{
			Id:                    cloudListenerInfo.Id,
			DefaultTlsContainerId: lo.ToPtr(cloudCertId),
		},
	}
	updateListenerResp, err := d.sdkClient.UpdateListener(updateListenerReq)
	d.logger.Debug("sdk request 'vlb.UpdateListener'", slog.Any("request", updateListenerReq), slog.Any("response", updateListenerResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'vlb.UpdateListener': %w", err)
	}

	return nil
}

func (d *Deployer) updateListenerSniCertificate(ctx context.Context, cloudListenerInfo model.ListLoadBalanceHTTPSListenerResponseContent, cloudCertId string) error {
	// 修改 HTTPS 监听器
	// REF: https://ecloud.10086.cn/op-help-center/doc/article/97024
	updateListenerReq := &model.UpdateListenerRequest{
		&model.UpdateListenerBody{
			Id:              cloudListenerInfo.Id,
			SniUp:           lo.ToPtr(true),
			SniContainerIds: append(cloudListenerInfo.SniContainerIdList, cloudCertId),
		},
	}
	updateListenerResp, err := d.sdkClient.UpdateListener(updateListenerReq)
	d.logger.Debug("sdk request 'vlb.UpdateListener'", slog.Any("request", updateListenerReq), slog.Any("response", updateListenerResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'vlb.UpdateListener': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, accessKeySecret, poolId string) (*ecloudsdkvlb.Client, error) {
	client := ecloudsdkvlb.NewClient(&config.Config{
		AccessKey: &accessKeyId,
		SecretKey: &accessKeySecret,
		PoolId:    &poolId,
	})

	return client, nil
}
