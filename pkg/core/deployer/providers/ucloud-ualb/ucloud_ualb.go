package ucloudualb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/lo"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/ucloud-ulb"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	ucloudsdk "github.com/certimate-go/certimate/pkg/sdk3rd/ucloud/ulb"
)

type DeployerConfig struct {
	// 优刻得 API 私钥。
	PrivateKey string `json:"privateKey"`
	// 优刻得 API 公钥。
	PublicKey string `json:"publicKey"`
	// 优刻得项目 ID。
	ProjectId string `json:"projectId,omitempty"`
	// 优刻得地域。
	Region string `json:"region"`
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// 负载均衡实例 ID。
	// 部署资源类型为 [RESOURCE_TYPE_LOADBALANCER]、[RESOURCE_TYPE_LISTENER] 时必填。
	LoadbalancerId string `json:"loadbalancerId,omitempty"`
	// 负载均衡监听器 ID。
	// 部署资源类型为 [RESOURCE_TYPE_LISTENER] 时必填。
	ListenerId string `json:"listenerId,omitempty"`
	// SNI 域名（不支持泛域名）。
	// 部署资源类型为 [RESOURCE_TYPE_LISTENER] 时选填。
	Domain string `json:"domain,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *ucloudsdk.ULBClient
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.PrivateKey, config.PublicKey, config.ProjectId, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		PrivateKey: config.PrivateKey,
		PublicKey:  config.PublicKey,
		ProjectId:  config.ProjectId,
		Region:     config.Region,
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

	// 获取 ALB 下的 HTTPS 监听器列表
	// REF: https://docs.ucloud.cn/api/ulb-api/describe_listeners
	listenerIds := make([]string, 0)
	describeListenersOffset := 0
	describeListenersLimit := 100
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		describeListenerReq := d.sdkClient.NewDescribeListenersRequest()
		describeListenerReq.LoadBalancerId = ucloud.String(d.config.LoadbalancerId)
		describeListenerReq.Offset = ucloud.Int(describeListenersOffset)
		describeListenerReq.Limit = ucloud.Int(describeListenersLimit)
		describeListenerResp, err := d.sdkClient.DescribeListeners(describeListenerReq)
		d.logger.Debug("sdk request 'ulb.DescribeListeners'", slog.Any("request", describeListenerReq), slog.Any("response", describeListenerResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'ulb.DescribeListeners': %w", err)
		}

		for _, listenerItem := range describeListenerResp.Listeners {
			if listenerItem.ListenerProtocol == "HTTPS" {
				listenerIds = append(listenerIds, listenerItem.ListenerId)
			}
		}

		if len(describeListenerResp.Listeners) < describeListenersLimit {
			break
		}

		describeListenersOffset += describeListenersLimit
	}

	// 遍历更新 Listener 证书
	if len(listenerIds) == 0 {
		d.logger.Info("no alb listeners to deploy")
	} else {
		d.logger.Info("found https listeners to deploy", slog.Any("listenerIds", listenerIds))
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

	if err := d.updateListenerCertificate(ctx, d.config.LoadbalancerId, d.config.ListenerId, cloudCertId); err != nil {
		return err
	}

	return nil
}

func (d *Deployer) updateListenerCertificate(ctx context.Context, cloudLoadbalancerId, cloudListenerId string, cloudCertId string) error {
	// 描述应用型负载均衡监听器
	// REF: https://docs.ucloud.cn/api/ulb-api/describe_listeners
	describeListenersReq := d.sdkClient.NewDescribeListenersRequest()
	describeListenersReq.LoadBalancerId = ucloud.String(cloudLoadbalancerId)
	describeListenersReq.ListenerId = ucloud.String(cloudListenerId)
	describeListenersReq.Limit = ucloud.Int(1)
	describeListenerResp, err := d.sdkClient.DescribeListeners(describeListenersReq)
	d.logger.Debug("sdk request 'ulb.DescribeListeners'", slog.Any("request", describeListenersReq), slog.Any("response", describeListenerResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'ulb.DescribeListeners': %w", err)
	} else if len(describeListenerResp.Listeners) == 0 {
		return fmt.Errorf("could not find listener '%s'", cloudListenerId)
	}

	// 跳过已部署过的监听器
	listenerInfo := describeListenerResp.Listeners[0]
	if d.config.Domain == "" {
		if lo.ContainsBy(listenerInfo.Certificates, func(item ulb.Certificate) bool { return item.SSLId == cloudCertId && item.IsDefault }) {
			return nil
		}
	} else {
		if lo.ContainsBy(listenerInfo.Certificates, func(item ulb.Certificate) bool { return item.SSLId == cloudCertId && !item.IsDefault }) {
			return nil
		}
	}

	if d.config.Domain == "" {
		// 未指定 SNI，只需部署到监听器

		updateListenerAttributeReq := d.sdkClient.NewUpdateListenerAttributeRequest()
		updateListenerAttributeReq.LoadBalancerId = ucloud.String(cloudLoadbalancerId)
		updateListenerAttributeReq.ListenerId = ucloud.String(cloudListenerId)
		updateListenerAttributeReq.Certificates = []string{cloudCertId}
		updateListenerResp, err := d.sdkClient.UpdateListenerAttribute(updateListenerAttributeReq)
		d.logger.Debug("sdk request 'ulb.UpdateListenerAttribute'", slog.Any("request", updateListenerAttributeReq), slog.Any("response", updateListenerResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'ulb.UpdateListenerAttribute': %w", err)
		}
	} else {
		// 指定 SNI，需部署到扩展域名

		// 新增监听器扩展证书
		// REF: https://docs.ucloud.cn/api/ulb-api/add_ssl_binding_json
		addSSLBindingReq := d.sdkClient.NewAddSSLBindingRequest()
		addSSLBindingReq.LoadBalancerId = ucloud.String(cloudLoadbalancerId)
		addSSLBindingReq.ListenerId = ucloud.String(cloudListenerId)
		addSSLBindingReq.SSLIds = []string{cloudCertId}
		addSSLBindingResp, err := d.sdkClient.AddSSLBinding(addSSLBindingReq)
		d.logger.Debug("sdk request 'ulb.AddSSLBinding'", slog.Any("request", addSSLBindingReq), slog.Any("response", addSSLBindingResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'ulb.AddSSLBinding': %w", err)
		}

		// 找出需要删除绑定的扩展证书
		// REF: https://docs.ucloud.cn/api/ulb-api/describe_sslv2
		sslIdsToDelete := make([]string, 0)
		for _, certItem := range listenerInfo.Certificates {
			if certItem.IsDefault {
				continue
			}

			describeSSLV2Req := d.sdkClient.NewDescribeSSLV2Request()
			describeSSLV2Req.SSLId = ucloud.String(certItem.SSLId)
			describeSSLV2Req.Limit = ucloud.Int(1)
			describeSSLV2Resp, err := d.sdkClient.DescribeSSLV2(describeSSLV2Req)
			d.logger.Debug("sdk request 'ulb.DescribeSSLV2'", slog.Any("request", describeSSLV2Req), slog.Any("response", describeSSLV2Resp))
			if err != nil {
				continue
			} else if len(describeSSLV2Resp.DataSet) == 0 {
				continue
			}

			sslItem := describeSSLV2Resp.DataSet[0]
			if sslItem.NotAfter != 0 && int64(sslItem.NotAfter) < time.Now().Unix() {
				sslIdsToDelete = append(sslIdsToDelete, sslItem.SSLId) // 过期证书需要删除
				continue
			} else if sslItem.Domains == d.config.Domain {
				sslIdsToDelete = append(sslIdsToDelete, sslItem.SSLId) // 同域名证书需要删除
				continue
			}
		}

		// 删除监听器绑定的扩展证书
		// REF: https://docs.ucloud.cn/api/ulb-api/delete_ssl_binding_json
		if len(sslIdsToDelete) > 0 {
			deleteSSLBindingReq := d.sdkClient.NewDeleteSSLBindingRequest()
			deleteSSLBindingReq.LoadBalancerId = ucloud.String(cloudLoadbalancerId)
			deleteSSLBindingReq.ListenerId = ucloud.String(cloudListenerId)
			deleteSSLBindingReq.SSLIds = sslIdsToDelete
			deleteSSLBindingResp, err := d.sdkClient.DeleteSSLBinding(deleteSSLBindingReq)
			d.logger.Debug("sdk request 'ulb.DeleteSSLBinding'", slog.Any("request", deleteSSLBindingReq), slog.Any("response", deleteSSLBindingResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'ulb.DeleteSSLBinding': %w", err)
			}
		}
	}

	return nil
}

func createSDKClient(privateKey, publicKey, projectId, region string) (*ucloudsdk.ULBClient, error) {
	if privateKey == "" {
		return nil, fmt.Errorf("ucloud: invalid private key")
	}
	if publicKey == "" {
		return nil, fmt.Errorf("ucloud: invalid public key")
	}

	cfg := ucloud.NewConfig()
	cfg.ProjectId = projectId
	cfg.Region = region

	credential := auth.NewCredential()
	credential.PrivateKey = privateKey
	credential.PublicKey = publicKey

	client := ucloudsdk.NewClient(&cfg, &credential)
	return client, nil
}
