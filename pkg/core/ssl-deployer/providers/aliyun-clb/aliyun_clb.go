package aliyunclb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	alislb "github.com/alibabacloud-go/slb-20140515/v4/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/certimate-go/certimate/pkg/core"
	sslmgrsp "github.com/certimate-go/certimate/pkg/core/ssl-manager/providers/aliyun-slb"
	"github.com/certimate-go/certimate/pkg/utils/ifelse"
)

type SSLDeployerProviderConfig struct {
	// 阿里云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 阿里云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 阿里云资源组 ID。
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
	// 阿里云地域。
	Region string `json:"region"`
	// 部署资源类型。
	ResourceType ResourceType `json:"resourceType"`
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

type SSLDeployerProvider struct {
	config     *SSLDeployerProviderConfig
	logger     *slog.Logger
	sdkClient  *alislb.Client
	sslManager core.SSLManager
}

var _ core.SSLDeployer = (*SSLDeployerProvider)(nil)

func NewSSLDeployerProvider(config *SSLDeployerProviderConfig) (*SSLDeployerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create sdk client: %w", err)
	}

	sslmgr, err := sslmgrsp.NewSSLManagerProvider(&sslmgrsp.SSLManagerProviderConfig{
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
		ResourceGroupId: config.ResourceGroupId,
		Region: ifelse.
			If[string](config.Region == "" || strings.HasPrefix(config.Region, "cn-")).
			Then("cn-hangzhou").
			Else("ap-southeast-1"),
	})
	if err != nil {
		return nil, fmt.Errorf("could not create ssl manager: %w", err)
	}

	return &SSLDeployerProvider{
		config:     config,
		logger:     slog.Default(),
		sdkClient:  client,
		sslManager: sslmgr,
	}, nil
}

func (d *SSLDeployerProvider) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}

	d.sslManager.SetLogger(logger)
}

func (d *SSLDeployerProvider) Deploy(ctx context.Context, certPEM string, privkeyPEM string) (*core.SSLDeployResult, error) {
	// 上传证书
	upres, err := d.sslManager.Upload(ctx, certPEM, privkeyPEM)
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

	return &core.SSLDeployResult{}, nil
}

func (d *SSLDeployerProvider) deployToLoadbalancer(ctx context.Context, cloudCertId string) error {
	if d.config.LoadbalancerId == "" {
		return errors.New("config `loadbalancerId` is required")
	}

	// 查询负载均衡实例的详细信息
	// REF: https://help.aliyun.com/zh/slb/classic-load-balancer/developer-reference/api-slb-2014-05-15-describeloadbalancerattribute
	describeLoadBalancerAttributeReq := &alislb.DescribeLoadBalancerAttributeRequest{
		RegionId:       tea.String(d.config.Region),
		LoadBalancerId: tea.String(d.config.LoadbalancerId),
	}
	describeLoadBalancerAttributeResp, err := d.sdkClient.DescribeLoadBalancerAttribute(describeLoadBalancerAttributeReq)
	d.logger.Debug("sdk request 'slb.DescribeLoadBalancerAttribute'", slog.Any("request", describeLoadBalancerAttributeReq), slog.Any("response", describeLoadBalancerAttributeResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'slb.DescribeLoadBalancerAttribute': %w", err)
	}

	// 查询 HTTPS 监听列表
	// REF: https://help.aliyun.com/zh/slb/classic-load-balancer/developer-reference/api-slb-2014-05-15-describeloadbalancerlisteners
	listenerPorts := make([]int32, 0)
	describeLoadBalancerListenersLimit := int32(100)
	var describeLoadBalancerListenersToken *string = nil
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		describeLoadBalancerListenersReq := &alislb.DescribeLoadBalancerListenersRequest{
			RegionId:         tea.String(d.config.Region),
			MaxResults:       tea.Int32(describeLoadBalancerListenersLimit),
			NextToken:        describeLoadBalancerListenersToken,
			LoadBalancerId:   []*string{tea.String(d.config.LoadbalancerId)},
			ListenerProtocol: tea.String("https"),
		}
		describeLoadBalancerListenersResp, err := d.sdkClient.DescribeLoadBalancerListeners(describeLoadBalancerListenersReq)
		d.logger.Debug("sdk request 'slb.DescribeLoadBalancerListeners'", slog.Any("request", describeLoadBalancerListenersReq), slog.Any("response", describeLoadBalancerListenersResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'slb.DescribeLoadBalancerListeners': %w", err)
		}

		if describeLoadBalancerListenersResp.Body.Listeners != nil {
			for _, listener := range describeLoadBalancerListenersResp.Body.Listeners {
				listenerPorts = append(listenerPorts, *listener.ListenerPort)
			}
		}

		if len(describeLoadBalancerListenersResp.Body.Listeners) == 0 || describeLoadBalancerListenersResp.Body.NextToken == nil {
			break
		} else {
			describeLoadBalancerListenersToken = describeLoadBalancerListenersResp.Body.NextToken
		}
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

func (d *SSLDeployerProvider) deployToListener(ctx context.Context, cloudCertId string) error {
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

func (d *SSLDeployerProvider) updateListenerCertificate(ctx context.Context, cloudLoadbalancerId string, cloudListenerPort int32, cloudCertId string) error {
	// 查询监听配置
	// REF: https://help.aliyun.com/zh/slb/classic-load-balancer/developer-reference/api-slb-2014-05-15-describeloadbalancerhttpslistenerattribute
	describeLoadBalancerHTTPSListenerAttributeReq := &alislb.DescribeLoadBalancerHTTPSListenerAttributeRequest{
		LoadBalancerId: tea.String(cloudLoadbalancerId),
		ListenerPort:   tea.Int32(cloudListenerPort),
	}
	describeLoadBalancerHTTPSListenerAttributeResp, err := d.sdkClient.DescribeLoadBalancerHTTPSListenerAttribute(describeLoadBalancerHTTPSListenerAttributeReq)
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
		setLoadBalancerHTTPSListenerAttributeResp, err := d.sdkClient.SetLoadBalancerHTTPSListenerAttribute(setLoadBalancerHTTPSListenerAttributeReq)
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
		describeDomainExtensionsResp, err := d.sdkClient.DescribeDomainExtensions(describeDomainExtensionsReq)
		d.logger.Debug("sdk request 'slb.DescribeDomainExtensions'", slog.Any("request", describeDomainExtensionsReq), slog.Any("response", describeDomainExtensionsResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'slb.DescribeDomainExtensions': %w", err)
		}

		// 遍历修改扩展域名
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
				setDomainExtensionAttributeResp, err := d.sdkClient.SetDomainExtensionAttribute(setDomainExtensionAttributeReq)
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

func createSDKClient(accessKeyId, accessKeySecret, region string) (*alislb.Client, error) {
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

	client, err := alislb.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
