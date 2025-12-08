package aliyunclb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	alislb "github.com/alibabacloud-go/slb-20140515/v4/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aliyun-slb"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-clb/internal"
)

type DeployerConfig struct {
	// 阿里云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 阿里云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 阿里云资源组 ID。
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
	// 阿里云地域。
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
	sdkClient  *internal.SlbClient
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
		ResourceGroupId: config.ResourceGroupId,
		Region:          config.Region,
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

	// 查询负载均衡实例的详细信息
	// REF: https://help.aliyun.com/zh/slb/classic-load-balancer/developer-reference/api-slb-2014-05-15-describeloadbalancerattribute
	describeLoadBalancerAttributeReq := &alislb.DescribeLoadBalancerAttributeRequest{
		RegionId:       tea.String(d.config.Region),
		LoadBalancerId: tea.String(d.config.LoadbalancerId),
	}
	describeLoadBalancerAttributeResp, err := d.sdkClient.DescribeLoadBalancerAttributeWithContext(ctx, describeLoadBalancerAttributeReq, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'slb.DescribeLoadBalancerAttribute'", slog.Any("request", describeLoadBalancerAttributeReq), slog.Any("response", describeLoadBalancerAttributeResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'slb.DescribeLoadBalancerAttribute': %w", err)
	}

	// 查询 HTTPS 监听列表
	// REF: https://help.aliyun.com/zh/slb/classic-load-balancer/developer-reference/api-slb-2014-05-15-describeloadbalancerlisteners
	listenerPorts := make([]int32, 0)
	describeLoadBalancerListenersToken := (*string)(nil)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		describeLoadBalancerListenersReq := &alislb.DescribeLoadBalancerListenersRequest{
			RegionId:         tea.String(d.config.Region),
			NextToken:        describeLoadBalancerListenersToken,
			MaxResults:       tea.Int32(100),
			LoadBalancerId:   tea.StringSlice([]string{d.config.LoadbalancerId}),
			ListenerProtocol: tea.String("https"),
		}
		describeLoadBalancerListenersResp, err := d.sdkClient.DescribeLoadBalancerListenersWithContext(ctx, describeLoadBalancerListenersReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'slb.DescribeLoadBalancerListeners'", slog.Any("request", describeLoadBalancerListenersReq), slog.Any("response", describeLoadBalancerListenersResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'slb.DescribeLoadBalancerListeners': %w", err)
		}

		if describeLoadBalancerListenersResp.Body == nil {
			break
		}

		for _, listener := range describeLoadBalancerListenersResp.Body.Listeners {
			listenerPorts = append(listenerPorts, *listener.ListenerPort)
		}

		if len(describeLoadBalancerListenersResp.Body.Listeners) == 0 || describeLoadBalancerListenersResp.Body.NextToken == nil {
			break
		}

		describeLoadBalancerListenersToken = describeLoadBalancerListenersResp.Body.NextToken
	}

	// 遍历更新监听证书
	if len(listenerPorts) == 0 {
		d.logger.Info("no clb listeners to deploy")
	} else {
		d.logger.Info("found https listeners to deploy", slog.Any("listenerPorts", listenerPorts))
		var errs []error

		for _, listenerPort := range listenerPorts {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := d.updateListenerCertificate(ctx, d.config.LoadbalancerId, listenerPort, cloudCertId); err != nil {
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

	// 更新监听
	if err := d.updateListenerCertificate(ctx, d.config.LoadbalancerId, d.config.ListenerPort, cloudCertId); err != nil {
		return err
	}

	return nil
}

func (d *Deployer) updateListenerCertificate(ctx context.Context, cloudLoadbalancerId string, cloudListenerPort int32, cloudCertId string) error {
	// 查询监听配置
	// REF: https://help.aliyun.com/zh/slb/classic-load-balancer/developer-reference/api-slb-2014-05-15-describeloadbalancerhttpslistenerattribute
	describeLoadBalancerHTTPSListenerAttributeReq := &alislb.DescribeLoadBalancerHTTPSListenerAttributeRequest{
		LoadBalancerId: tea.String(cloudLoadbalancerId),
		ListenerPort:   tea.Int32(cloudListenerPort),
	}
	describeLoadBalancerHTTPSListenerAttributeResp, err := d.sdkClient.DescribeLoadBalancerHTTPSListenerAttributeWithContext(ctx, describeLoadBalancerHTTPSListenerAttributeReq, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'slb.DescribeLoadBalancerHTTPSListenerAttribute'", slog.Any("request", describeLoadBalancerHTTPSListenerAttributeReq), slog.Any("response", describeLoadBalancerHTTPSListenerAttributeResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'slb.DescribeLoadBalancerHTTPSListenerAttribute': %w", err)
	}

	if d.config.Domain == "" {
		// 未指定 SNI，只需部署到监听器

		// 修改监听配置
		// REF: https://help.aliyun.com/zh/slb/classic-load-balancer/developer-reference/api-slb-2014-05-15-setloadbalancerhttpslistenerattribute
		setLoadBalancerHTTPSListenerAttributeReq := &alislb.SetLoadBalancerHTTPSListenerAttributeRequest{
			RegionId:            tea.String(d.config.Region),
			LoadBalancerId:      tea.String(cloudLoadbalancerId),
			ListenerPort:        tea.Int32(cloudListenerPort),
			ServerCertificateId: tea.String(cloudCertId),
		}
		setLoadBalancerHTTPSListenerAttributeResp, err := d.sdkClient.SetLoadBalancerHTTPSListenerAttributeWithContext(ctx, setLoadBalancerHTTPSListenerAttributeReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'slb.SetLoadBalancerHTTPSListenerAttribute'", slog.Any("request", setLoadBalancerHTTPSListenerAttributeReq), slog.Any("response", setLoadBalancerHTTPSListenerAttributeResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'slb.SetLoadBalancerHTTPSListenerAttribute': %w", err)
		}
	} else {
		// 指定 SNI，需部署到扩展域名

		// 查询扩展域名
		// REF: https://help.aliyun.com/zh/slb/classic-load-balancer/developer-reference/api-slb-2014-05-15-describedomainextensions
		describeDomainExtensionsReq := &alislb.DescribeDomainExtensionsRequest{
			RegionId:       tea.String(d.config.Region),
			LoadBalancerId: tea.String(cloudLoadbalancerId),
			ListenerPort:   tea.Int32(cloudListenerPort),
		}
		describeDomainExtensionsResp, err := d.sdkClient.DescribeDomainExtensionsWithContext(ctx, describeDomainExtensionsReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'slb.DescribeDomainExtensions'", slog.Any("request", describeDomainExtensionsReq), slog.Any("response", describeDomainExtensionsResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'slb.DescribeDomainExtensions': %w", err)
		}

		// 遍历修改扩展域名证书
		// REF: https://help.aliyun.com/zh/slb/classic-load-balancer/developer-reference/api-slb-2014-05-15-setdomainextensionattribute
		if describeDomainExtensionsResp.Body.DomainExtensions != nil && describeDomainExtensionsResp.Body.DomainExtensions.DomainExtension != nil {
			var errs []error

			for _, domainExtension := range describeDomainExtensionsResp.Body.DomainExtensions.DomainExtension {
				if *domainExtension.Domain != d.config.Domain {
					continue
				}

				setDomainExtensionAttributeReq := &alislb.SetDomainExtensionAttributeRequest{
					RegionId:            tea.String(d.config.Region),
					DomainExtensionId:   tea.String(*domainExtension.DomainExtensionId),
					ServerCertificateId: tea.String(cloudCertId),
				}
				setDomainExtensionAttributeResp, err := d.sdkClient.SetDomainExtensionAttributeWithContext(ctx, setDomainExtensionAttributeReq, &dara.RuntimeOptions{})
				d.logger.Debug("sdk request 'slb.SetDomainExtensionAttribute'", slog.Any("request", setDomainExtensionAttributeReq), slog.Any("response", setDomainExtensionAttributeResp))
				if err != nil {
					errs = append(errs, fmt.Errorf("failed to execute sdk request 'slb.SetDomainExtensionAttribute': %w", err))
					continue
				}
			}

			if len(errs) > 0 {
				return errors.Join(errs...)
			}
		}
	}

	return nil
}

func createSDKClient(accessKeyId, accessKeySecret, region string) (*internal.SlbClient, error) {
	// 接入点一览 https://api.aliyun.com/product/Slb
	var endpoint string
	switch region {
	case "",
		"cn-hangzhou",
		"cn-hangzhou-finance",
		"cn-shanghai-finance-1",
		"cn-shenzhen-finance-1":
		endpoint = "slb.aliyuncs.com"
	default:
		endpoint = fmt.Sprintf("slb.%s.aliyuncs.com", region)
	}

	config := &aliopen.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(endpoint),
	}

	client, err := internal.NewSlbClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
