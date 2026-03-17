package ctcccloudelb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/ctcccloud-elb"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	ctyunelb "github.com/certimate-go/certimate/pkg/sdk3rd/ctyun/elb"
)

type DeployerConfig struct {
	// 天翼云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 天翼云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 天翼云资源池 ID。
	RegionId string `json:"regionId"`
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// 负载均衡实例 ID。
	// 部署资源类型为 [RESOURCE_TYPE_LOADBALANCER] 时必填。
	LoadbalancerId string `json:"loadbalancerId,omitempty"`
	// 负载均衡监听器 ID。
	// 部署资源类型为 [RESOURCE_TYPE_LISTENER] 时必填。
	ListenerId string `json:"listenerId,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *ctyunelb.Client
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		SecretAccessKey: config.SecretAccessKey,
		RegionId:        config.RegionId,
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
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
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

	// 查询监听列表
	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=24&api=5654&data=88&isNormal=1&vid=82
	listenerIds := make([]string, 0)
	{
		listListenersReq := &ctyunelb.ListListenersRequest{
			RegionID:       lo.ToPtr(d.config.RegionId),
			LoadBalancerID: lo.ToPtr(d.config.LoadbalancerId),
		}
		listListenersResp, err := d.sdkClient.ListListenersWithContext(ctx, listListenersReq)
		d.logger.Debug("sdk request 'elb.ListListeners'", slog.Any("request", listListenersReq), slog.Any("response", listListenersResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'elb.ListListeners': %w", err)
		}

		for _, listener := range listListenersResp.ReturnObj {
			if strings.EqualFold(listener.Protocol, "HTTPS") {
				listenerIds = append(listenerIds, listener.ID)
			}
		}
	}

	// 遍历更新监听证书
	if len(listenerIds) == 0 {
		d.logger.Info("no elb listeners to deploy")
	} else {
		d.logger.Info("found https listeners to deploy", slog.Any("listenerIds", listenerIds))
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

	// 更新监听
	if err := d.updateListenerCertificate(ctx, d.config.ListenerId, cloudCertId); err != nil {
		return err
	}

	return nil
}

func (d *Deployer) updateListenerCertificate(ctx context.Context, cloudListenerId string, cloudCertId string) error {
	// 更新监听器
	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=24&api=5652&data=88&isNormal=1&vid=82
	setLoadBalancerHTTPSListenerAttributeReq := &ctyunelb.UpdateListenerRequest{
		RegionID:      lo.ToPtr(d.config.RegionId),
		ListenerID:    lo.ToPtr(cloudListenerId),
		CertificateID: lo.ToPtr(cloudCertId),
	}
	setLoadBalancerHTTPSListenerAttributeResp, err := d.sdkClient.UpdateListenerWithContext(ctx, setLoadBalancerHTTPSListenerAttributeReq)
	d.logger.Debug("sdk request 'elb.UpdateListener'", slog.Any("request", setLoadBalancerHTTPSListenerAttributeReq), slog.Any("response", setLoadBalancerHTTPSListenerAttributeResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'elb.UpdateListener': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, secretAccessKey string) (*ctyunelb.Client, error) {
	return ctyunelb.NewClient(accessKeyId, secretAccessKey)
}
