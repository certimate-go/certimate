package aliyunclb

import (
	"context"
	"fmt"
	"log/slog"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"

	alislb "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/alibabacloud-go/slb-20140515/v4/client"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aliyun-slb"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
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
	// 部署目标。
	DeployTarget string `json:"deployTarget"`
	// 负载均衡实例 ID。
	// 部署目标为 [DEPLOY_TARGET_LOADBALANCER]、[DEPLOY_TARGET_LISTENER] 时必填。
	LoadbalancerId string `json:"loadbalancerId,omitempty"`
	// 负载均衡监听端口。
	// 部署目标为 [DEPLOY_TARGET_LISTENER] 时必填。
	ListenerPort int32 `json:"listenerPort,omitempty"`
	// SNI 域名（支持泛域名）。
	// 部署目标为 [DEPLOY_TARGET_LOADBALANCER]、[DEPLOY_TARGET_LISTENER] 时选填。
	Domain string `json:"domain,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *alislb.Client
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
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

	// 批量更新监听证书
	if len(listenerPorts) == 0 {
		d.logger.Info("no clb listeners to deploy")
	} else {
		d.logger.Info("found clb listeners to deploy", slog.Any("listenerPorts", listenerPorts))

		if err := xloop.ForRangeAllWithContext(ctx, listenerPorts, func(ctx context.Context, listenerPort int32, _ int) error {
			return d.updateListenerCertificate(ctx, d.config.LoadbalancerId, listenerPort, cloudCertId)
		}); err != nil {
			return err
		}
	}

	return nil
}

func (d *Deployer) deployToListener(ctx context.Context, cloudCertId string) error {
	if d.config.LoadbalancerId == "" {
		return fmt.Errorf("config `loadbalancerId` is required")
	}
	if d.config.ListenerPort == 0 {
		return fmt.Errorf("config `listenerPort` is required")
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
		if tea.StringValue(describeLoadBalancerHTTPSListenerAttributeResp.Body.ServerCertificateId) == cloudCertId {
			d.logger.Info("no need to deploy clb listener default certificate")
			return nil
		}
		return d.updateListenerDefaultCertificate(ctx, cloudLoadbalancerId, cloudListenerPort, cloudCertId)
	} else {
		// 指定 SNI，需部署到扩展域名
		return d.updateListenerSniCertificate(ctx, cloudLoadbalancerId, cloudListenerPort, cloudCertId)
	}
}

func (d *Deployer) updateListenerDefaultCertificate(ctx context.Context, cloudLoadbalancerId string, cloudListenerPort int32, cloudCertId string) error {
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

	return nil
}

func (d *Deployer) updateListenerSniCertificate(ctx context.Context, cloudLoadbalancerId string, cloudListenerPort int32, cloudCertId string) error {
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

	// 修改扩展域名证书
	// REF: https://help.aliyun.com/zh/slb/classic-load-balancer/developer-reference/api-slb-2014-05-15-setdomainextensionattribute
	if describeDomainExtensionsResp.Body.DomainExtensions != nil && describeDomainExtensionsResp.Body.DomainExtensions.DomainExtension != nil {
		domainExtensionIds := make([]string, 0)
		for _, domainExtension := range describeDomainExtensionsResp.Body.DomainExtensions.DomainExtension {
			if tea.StringValue(domainExtension.Domain) != d.config.Domain {
				continue
			}

			if tea.StringValue(domainExtension.ServerCertificateId) == cloudCertId {
				continue
			}

			domainExtensionIds = append(domainExtensionIds, tea.StringValue(domainExtension.DomainExtensionId))
		}

		if err := xloop.ForRangeAllWithContext(ctx, domainExtensionIds, func(ctx context.Context, domainExtensionId string, _ int) error {
			setDomainExtensionAttributeReq := &alislb.SetDomainExtensionAttributeRequest{
				RegionId:            tea.String(d.config.Region),
				DomainExtensionId:   tea.String(domainExtensionId),
				ServerCertificateId: tea.String(cloudCertId),
			}
			setDomainExtensionAttributeResp, err := d.sdkClient.SetDomainExtensionAttributeWithContext(ctx, setDomainExtensionAttributeReq, &dara.RuntimeOptions{})
			d.logger.Debug("sdk request 'slb.SetDomainExtensionAttribute'", slog.Any("request", setDomainExtensionAttributeReq), slog.Any("response", setDomainExtensionAttributeResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'slb.SetDomainExtensionAttribute': %w", err)
			}

			return nil
		}); err != nil {
			return err
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
