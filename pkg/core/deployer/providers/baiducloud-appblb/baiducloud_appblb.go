package baiducloudappblb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	bceappblb "github.com/baidubce/bce-sdk-go/services/appblb"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/baiducloud-cert"
	"github.com/certimate-go/certimate/pkg/core/deployer"
)

type DeployerConfig struct {
	// 百度智能云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 百度智能云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 百度智能云区域。
	Region string `json:"region"`
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// 负载均衡实例 ID。
	// 部署资源类型为 [RESOURCE_TYPE_LOADBALANCER]、[RESOURCE_TYPE_LISTENER] 时必填。
	LoadbalancerId string `json:"loadbalancerId,omitempty"`
	// 负载均衡监听端口。
	// 部署资源类型为 [RESOURCE_TYPE_LISTENER] 时必填。
	ListenerPort int32 `json:"listenerPort,omitempty"`
	// SNI 域名（支持泛域名）。
	// 部署资源类型为 [RESOURCE_TYPE_LOADBALANCER]、[RESOURCE_TYPE_LISTENER] 时选填。
	Domain string `json:"domain,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *bceappblb.Client
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		SecretAccessKey: config.SecretAccessKey,
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

	// 查询 BLB 实例详情
	// REF: https://cloud.baidu.com/doc/BLB/s/6jwvxnyhi#describeloadbalancerdetail%E6%9F%A5%E8%AF%A2blb%E5%AE%9E%E4%BE%8B%E8%AF%A6%E6%83%85
	describeLoadBalancerDetailResp, err := d.sdkClient.DescribeLoadBalancerDetail(d.config.LoadbalancerId)
	d.logger.Debug("sdk request 'appblb.DescribeLoadBalancerAttribute'", slog.String("blbId", d.config.LoadbalancerId), slog.Any("response", describeLoadBalancerDetailResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'appblb.DescribeLoadBalancerDetail': %w", err)
	}

	// 获取全部 HTTPS/SSL 监听端口
	listeners := make([]struct {
		Type string
		Port int32
	}, 0)
	for _, listener := range describeLoadBalancerDetailResp.Listener {
		if listener.Type == "HTTPS" || listener.Type == "SSL" {
			listenerPort, err := strconv.Atoi(listener.Port)
			if err != nil {
				continue
			}

			listeners = append(listeners, struct {
				Type string
				Port int32
			}{
				Type: listener.Type,
				Port: int32(listenerPort),
			})
		}
	}

	// 遍历更新监听证书
	if len(listeners) == 0 {
		d.logger.Info("no blb listeners to deploy")
	} else {
		d.logger.Info("found https/ssl listeners to deploy", slog.Any("listeners", listeners))
		var errs []error

		for _, listener := range listeners {
			select {
			case <-ctx.Done():
				return ctx.Err()

			default:
				if err := d.updateListenerCertificate(ctx, d.config.LoadbalancerId, listener.Type, listener.Port, cloudCertId); err != nil {
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
	if d.config.ListenerPort == 0 {
		return errors.New("config `listenerPort` is required")
	}

	// 查询监听
	// REF: https://cloud.baidu.com/doc/BLB/s/ujwvxnyux#describeappalllisteners%E6%9F%A5%E8%AF%A2%E6%89%80%E6%9C%89%E7%9B%91%E5%90%AC
	describeAppAllListenersRequest := &bceappblb.DescribeAppListenerArgs{
		ListenerPort: uint16(d.config.ListenerPort),
	}
	describeAppAllListenersResp, err := d.sdkClient.DescribeAppAllListeners(d.config.LoadbalancerId, describeAppAllListenersRequest)
	d.logger.Debug("sdk request 'appblb.DescribeAppAllListeners'", slog.String("blbId", d.config.LoadbalancerId), slog.Any("request", describeAppAllListenersRequest), slog.Any("response", describeAppAllListenersResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'appblb.DescribeAppAllListeners': %w", err)
	}

	// 获取全部 HTTPS/SSL 监听端口
	listeners := make([]struct {
		Type string
		Port int32
	}, 0)
	for _, listener := range describeAppAllListenersResp.ListenerList {
		if listener.ListenerType == "HTTPS" || listener.ListenerType == "SSL" {
			listeners = append(listeners, struct {
				Type string
				Port int32
			}{
				Type: listener.ListenerType,
				Port: int32(listener.ListenerPort),
			})
		}
	}

	// 遍历更新监听证书
	if len(listeners) == 0 {
		d.logger.Info("no blb listeners to deploy")
	} else {
		d.logger.Info("found https/ssl listeners to deploy", slog.Any("listeners", listeners))
		var errs []error

		for _, listener := range listeners {
			select {
			case <-ctx.Done():
				return ctx.Err()

			default:
				if err := d.updateListenerCertificate(ctx, d.config.LoadbalancerId, listener.Type, listener.Port, cloudCertId); err != nil {
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

func (d *Deployer) updateListenerCertificate(ctx context.Context, cloudLoadbalancerId string, cloudListenerType string, cloudListenerPort int32, cloudCertId string) error {
	switch strings.ToUpper(cloudListenerType) {
	case "HTTPS":
		return d.updateHttpsListenerCertificate(ctx, cloudLoadbalancerId, cloudListenerPort, cloudCertId)
	case "SSL":
		return d.updateSslListenerCertificate(ctx, cloudLoadbalancerId, cloudListenerPort, cloudCertId)
	default:
		return fmt.Errorf("unsupported listener type '%s'", cloudListenerType)
	}
}

func (d *Deployer) updateHttpsListenerCertificate(ctx context.Context, cloudLoadbalancerId string, cloudHttpsListenerPort int32, cloudCertId string) error {
	// 查询 HTTPS 监听器
	// REF: https://cloud.baidu.com/doc/BLB/s/ujwvxnyux#describeapphttpslisteners%E6%9F%A5%E8%AF%A2https%E7%9B%91%E5%90%AC%E5%99%A8
	describeAppHTTPSListenersReq := &bceappblb.DescribeAppListenerArgs{
		ListenerPort: uint16(cloudHttpsListenerPort),
		MaxKeys:      1,
	}
	describeAppHTTPSListenersResp, err := d.sdkClient.DescribeAppHTTPSListeners(cloudLoadbalancerId, describeAppHTTPSListenersReq)
	d.logger.Debug("sdk request 'appblb.DescribeAppHTTPSListeners'", slog.String("blbId", cloudLoadbalancerId), slog.Any("request", describeAppHTTPSListenersReq), slog.Any("response", describeAppHTTPSListenersResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'appblb.DescribeAppHTTPSListeners': %w", err)
	} else if len(describeAppHTTPSListenersResp.ListenerList) == 0 {
		return fmt.Errorf("could not find listener '%s:%d'", cloudLoadbalancerId, cloudHttpsListenerPort)
	}

	if d.config.Domain == "" {
		// 未指定 SNI，只需部署到监听器

		// 更新 HTTPS 监听器
		// REF: https://cloud.baidu.com/doc/BLB/s/ujwvxnyux#updateapphttpslistener%E6%9B%B4%E6%96%B0https%E7%9B%91%E5%90%AC%E5%99%A8
		updateAppHTTPSListenerReq := &bceappblb.UpdateAppHTTPSListenerArgs{
			ClientToken:  security.RandomString(32),
			ListenerPort: uint16(cloudHttpsListenerPort),
			Scheduler:    describeAppHTTPSListenersResp.ListenerList[0].Scheduler,
			CertIds:      []string{cloudCertId},
		}
		err := d.sdkClient.UpdateAppHTTPSListener(cloudLoadbalancerId, updateAppHTTPSListenerReq)
		d.logger.Debug("sdk request 'appblb.UpdateAppHTTPSListener'", slog.Any("request", updateAppHTTPSListenerReq))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'appblb.UpdateAppHTTPSListener': %w", err)
		}
	} else {
		// 指定 SNI，需部署到扩展域名

		// 更新 HTTPS 监听器
		// REF: https://cloud.baidu.com/doc/BLB/s/yjwvxnvl6#updatehttpslistener%E6%9B%B4%E6%96%B0https%E7%9B%91%E5%90%AC%E5%99%A8
		updateAppHTTPSListenerReq := &bceappblb.UpdateAppHTTPSListenerArgs{
			ClientToken:  security.RandomString(32),
			ListenerPort: uint16(cloudHttpsListenerPort),
			Scheduler:    describeAppHTTPSListenersResp.ListenerList[0].Scheduler,
			CertIds:      describeAppHTTPSListenersResp.ListenerList[0].CertIds,
			AdditionalCertDomains: lo.Map(describeAppHTTPSListenersResp.ListenerList[0].AdditionalCertDomains, func(domain bceappblb.AdditionalCertDomainsModel, _ int) bceappblb.AdditionalCertDomainsModel {
				if domain.Host == d.config.Domain {
					return bceappblb.AdditionalCertDomainsModel{
						Host:   domain.Host,
						CertId: cloudCertId,
					}
				}

				return bceappblb.AdditionalCertDomainsModel{
					Host:   domain.Host,
					CertId: domain.CertId,
				}
			}),
		}
		err := d.sdkClient.UpdateAppHTTPSListener(cloudLoadbalancerId, updateAppHTTPSListenerReq)
		d.logger.Debug("sdk request 'appblb.UpdateAppHTTPSListener'", slog.Any("request", updateAppHTTPSListenerReq))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'appblb.UpdateAppHTTPSListener': %w", err)
		}
	}

	return nil
}

func (d *Deployer) updateSslListenerCertificate(ctx context.Context, cloudLoadbalancerId string, cloudHttpsListenerPort int32, cloudCertId string) error {
	// 更新 SSL 监听器
	// REF: https://cloud.baidu.com/doc/BLB/s/ujwvxnyux#updateappssllistener%E6%9B%B4%E6%96%B0ssl%E7%9B%91%E5%90%AC%E5%99%A8
	updateAppSSLListenerReq := &bceappblb.UpdateAppSSLListenerArgs{
		ClientToken:  security.RandomString(32),
		ListenerPort: uint16(cloudHttpsListenerPort),
		CertIds:      []string{cloudCertId},
	}
	err := d.sdkClient.UpdateAppSSLListener(cloudLoadbalancerId, updateAppSSLListenerReq)
	d.logger.Debug("sdk request 'appblb.UpdateAppSSLListener'", slog.Any("request", updateAppSSLListenerReq))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'appblb.UpdateAppSSLListener': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*bceappblb.Client, error) {
	endpoint := ""
	if region != "" {
		endpoint = fmt.Sprintf("blb.%s.baidubce.com", region)
	}

	client, err := bceappblb.NewClient(accessKeyId, secretAccessKey, endpoint)
	if err != nil {
		return nil, err
	}

	return client, nil
}
