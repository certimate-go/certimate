package aliyunalb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/samber/lo"

	alialb "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/alibabacloud-go/alb-20200616/v2/client"
	alicas "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/alibabacloud-go/cas-20200407/v4/client"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aliyun-cas"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
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
	// 部署目标为 [DEPLOY_TARGET_LOADBALANCER] 时必填。
	LoadbalancerId string `json:"loadbalancerId,omitempty"`
	// 负载均衡监听 ID。
	// 部署目标为 [DEPLOY_TARGET_LISTENER] 时必填。
	ListenerId string `json:"listenerId,omitempty"`
	// SNI 域名（支持泛域名）。
	// 部署目标为 [DEPLOY_TARGET_LOADBALANCER]、[DEPLOY_TARGET_LISTENER] 时选填。
	Domain string `json:"domain,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClients *wSDKClients
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

type wSDKClients struct {
	ALB *alialb.Client
	CAS *alicas.Client
}

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	clients, err := createSDKClients(config.AccessKeyId, config.AccessKeySecret, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
		ResourceGroupId: config.ResourceGroupId,
		Region: lo.
			If(config.Region == "" || strings.HasPrefix(config.Region, "cn-"), "cn-hangzhou").
			Else("ap-southeast-1"),
	})
	if err != nil {
		return nil, fmt.Errorf("could not create certmgr: %w", err)
	}

	return &Deployer{
		config:     config,
		logger:     slog.Default(),
		sdkClients: clients,
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
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

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
		if err := d.deployToLoadbalancer(ctx, upres.ExtendedData["CertIdentifier"].(string), certX509.DNSNames); err != nil {
			return nil, err
		}

	case DEPLOY_TARGET_LISTENER:
		if err := d.deployToListener(ctx, upres.ExtendedData["CertIdentifier"].(string), certX509.DNSNames); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported deploy target '%s'", d.config.DeployTarget)
	}

	return &DeployResult{}, nil
}

func (d *Deployer) deployToLoadbalancer(ctx context.Context, cloudCertId string, cloudCertSANs []string) error {
	if d.config.LoadbalancerId == "" {
		return fmt.Errorf("config `loadbalancerId` is required")
	}

	// 查询负载均衡实例的详细信息
	// REF: https://help.aliyun.com/zh/slb/application-load-balancer/developer-reference/api-alb-2020-06-16-getloadbalancerattribute
	getLoadBalancerAttributeReq := &alialb.GetLoadBalancerAttributeRequest{
		LoadBalancerId: tea.String(d.config.LoadbalancerId),
	}
	getLoadBalancerAttributeResp, err := d.sdkClients.ALB.GetLoadBalancerAttributeWithContext(ctx, getLoadBalancerAttributeReq, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'alb.GetLoadBalancerAttribute'", slog.Any("request", getLoadBalancerAttributeReq), slog.Any("response", getLoadBalancerAttributeResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'alb.GetLoadBalancerAttribute': %w", err)
	}

	// 查询 HTTPS 监听列表
	// REF: https://help.aliyun.com/zh/slb/application-load-balancer/developer-reference/api-alb-2020-06-16-listlisteners
	listenerIds := make([]string, 0)
	listListenersToken := (*string)(nil)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		listListenersReq := &alialb.ListListenersRequest{
			NextToken:        listListenersToken,
			MaxResults:       tea.Int32(100),
			LoadBalancerIds:  tea.StringSlice([]string{d.config.LoadbalancerId}),
			ListenerProtocol: tea.String("HTTPS"),
		}
		listListenersResp, err := d.sdkClients.ALB.ListListenersWithContext(ctx, listListenersReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'alb.ListListeners'", slog.Any("request", listListenersReq), slog.Any("response", listListenersResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'alb.ListListeners': %w", err)
		}

		if listListenersResp.Body == nil {
			break
		}

		for _, listener := range listListenersResp.Body.Listeners {
			listenerIds = append(listenerIds, tea.StringValue(listener.ListenerId))
		}

		if len(listListenersResp.Body.Listeners) == 0 || listListenersResp.Body.NextToken == nil {
			break
		}

		listListenersToken = listListenersResp.Body.NextToken
	}

	// 查询 QUIC 监听列表
	// REF: https://help.aliyun.com/zh/slb/application-load-balancer/developer-reference/api-alb-2020-06-16-listlisteners
	listListenersToken = nil
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		listListenersReq := &alialb.ListListenersRequest{
			NextToken:        listListenersToken,
			MaxResults:       tea.Int32(100),
			LoadBalancerIds:  tea.StringSlice([]string{d.config.LoadbalancerId}),
			ListenerProtocol: tea.String("QUIC"),
		}
		listListenersResp, err := d.sdkClients.ALB.ListListenersWithContext(ctx, listListenersReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'alb.ListListeners'", slog.Any("request", listListenersReq), slog.Any("response", listListenersResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'alb.ListListeners': %w", err)
		}

		if listListenersResp.Body == nil {
			break
		}

		for _, listener := range listListenersResp.Body.Listeners {
			listenerIds = append(listenerIds, tea.StringValue(listener.ListenerId))
		}

		if len(listListenersResp.Body.Listeners) == 0 || listListenersResp.Body.NextToken == nil {
			break
		}

		listListenersToken = listListenersResp.Body.NextToken
	}

	// 遍历更新监听证书
	if len(listenerIds) == 0 {
		d.logger.Info("no alb listeners to deploy")
	} else {
		var errs []error
		d.logger.Info("found https/quic listeners to deploy", slog.Any("listenerIds", listenerIds))

		for _, listenerId := range listenerIds {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := d.updateListenerCertificate(ctx, listenerId, cloudCertId, cloudCertSANs); err != nil {
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

func (d *Deployer) deployToListener(ctx context.Context, cloudCertId string, cloudCertSANs []string) error {
	if d.config.ListenerId == "" {
		return fmt.Errorf("config `listenerId` is required")
	}

	// 更新监听
	if err := d.updateListenerCertificate(ctx, d.config.ListenerId, cloudCertId, cloudCertSANs); err != nil {
		return err
	}

	return nil
}

func (d *Deployer) updateListenerCertificate(ctx context.Context, cloudListenerId string, cloudCertId string, cloudCertSANs []string) error {
	if d.config.Domain == "" {
		// 未指定 SNI，只需部署到监听器

		if err := d.waitForListenerReady(ctx, cloudListenerId); err != nil {
			return err
		}

		// 修改监听的属性
		// REF: https://help.aliyun.com/zh/slb/application-load-balancer/developer-reference/api-alb-2020-06-16-updatelistenerattribute
		updateListenerAttributeReq := &alialb.UpdateListenerAttributeRequest{
			ListenerId: tea.String(cloudListenerId),
			Certificates: []*alialb.UpdateListenerAttributeRequestCertificates{{
				CertificateId: tea.String(cloudCertId),
			}},
		}
		updateListenerAttributeResp, err := d.sdkClients.ALB.UpdateListenerAttributeWithContext(ctx, updateListenerAttributeReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'alb.UpdateListenerAttribute'", slog.Any("request", updateListenerAttributeReq), slog.Any("response", updateListenerAttributeResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'alb.UpdateListenerAttribute': %w", err)
		}
	} else {
		// 指定 SNI，需部署到扩展域名

		// 查询监听证书列表
		// REF: https://help.aliyun.com/zh/slb/application-load-balancer/developer-reference/api-alb-2020-06-16-listlistenercertificates
		listenerExtCertificates := make([]alialb.ListListenerCertificatesResponseBodyCertificates, 0)
		listListenerCertificatesToken := (*string)(nil)
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			listListenerCertificatesReq := &alialb.ListListenerCertificatesRequest{
				NextToken:       listListenerCertificatesToken,
				MaxResults:      tea.Int32(100),
				ListenerId:      tea.String(cloudListenerId),
				CertificateType: tea.String("Server"),
			}
			listListenerCertificatesResp, err := d.sdkClients.ALB.ListListenerCertificatesWithContext(ctx, listListenerCertificatesReq, &dara.RuntimeOptions{})
			d.logger.Debug("sdk request 'alb.ListListenerCertificates'", slog.Any("request", listListenerCertificatesReq), slog.Any("response", listListenerCertificatesResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'alb.ListListenerCertificates': %w", err)
			}

			if listListenerCertificatesResp.Body == nil {
				break
			}

			for _, certItem := range listListenerCertificatesResp.Body.Certificates {
				if tea.BoolValue(certItem.IsDefault) {
					continue
				}

				if !strings.EqualFold(tea.StringValue(certItem.CertificateType), "Server") {
					continue
				}

				if !strings.EqualFold(tea.StringValue(certItem.Status), "Associated") {
					continue
				}

				listenerExtCertificates = append(listenerExtCertificates, *certItem)
			}

			if len(listListenerCertificatesResp.Body.Certificates) == 0 || listListenerCertificatesResp.Body.NextToken == nil {
				break
			}

			listListenerCertificatesToken = listListenerCertificatesResp.Body.NextToken
		}

		// 查询监听证书，并找出需要解除关联的证书
		// REF: https://help.aliyun.com/zh/slb/application-load-balancer/developer-reference/api-alb-2020-06-16-listlistenercertificates
		// REF: https://help.aliyun.com/zh/ssl-certificate/developer-reference/api-cas-2020-04-07-getcertificatedetail
		certificateIsAlreadyAssociated := false
		certificateIdsToDissociate := make([]string, 0)
		if len(listenerExtCertificates) > 0 {
			d.logger.Info("found listener certificates in used", slog.Any("certificates", listenerExtCertificates))
			var errs []error

			for _, listenerCertificate := range listenerExtCertificates {
				certIdWithRegion := tea.StringValue(listenerCertificate.CertificateId)
				if certIdWithRegion == cloudCertId {
					certificateIsAlreadyAssociated = true
					break
				}

				certIdBare := strings.SplitN(certIdWithRegion, "-", 2)[0]
				certIdBareAsInt64, err := strconv.ParseInt(certIdBare, 10, 64)
				if err != nil {
					errs = append(errs, err)
					continue
				}

				getCertificateDetailReq := &alicas.GetCertificateDetailRequest{
					CertificateId: tea.Int64(certIdBareAsInt64),
				}
				getCertificateDetailResp, err := d.sdkClients.CAS.GetCertificateDetailWithContext(ctx, getCertificateDetailReq, &dara.RuntimeOptions{})
				d.logger.Debug("sdk request 'cas.GetCertificateDetail'", slog.Any("request", getCertificateDetailReq), slog.Any("response", getCertificateDetailResp))
				if err != nil {
					if sdkErr, ok := err.(*tea.SDKError); ok {
						if sdkErrCode := tea.StringValue(sdkErr.Code); strings.HasPrefix(sdkErrCode, "NotFound") {
							continue
						}
					}

					errs = append(errs, fmt.Errorf("failed to execute sdk request 'cas.GetCertificateDetail': %w", err))
					continue
				} else {
					// 注意，虽然文档中存在 SubjectAlternativeNames 字段，但实际返回的数据结构中不包含
					certSANMatched := lo.ElementsMatch(strings.Split(tea.StringValue(getCertificateDetailResp.Body.Domain), ","), cloudCertSANs)
					if certSANMatched && lo.Contains(cloudCertSANs, d.config.Domain) { // 同域名证书需要删除
						certificateIdsToDissociate = append(certificateIdsToDissociate, certIdWithRegion)
						continue
					}

					certNotAfter := time.Unix(tea.Int64Value(getCertificateDetailResp.Body.NotAfter)/1000, 0)
					if !certNotAfter.IsZero() && certNotAfter.Before(time.Now()) { // 过期证书需要删除
						certificateIdsToDissociate = append(certificateIdsToDissociate, certIdWithRegion)
						continue
					}
				}
			}

			if len(errs) > 0 {
				return errors.Join(errs...)
			}
		}

		// 关联监听和扩展证书
		// REF: https://help.aliyun.com/zh/slb/application-load-balancer/developer-reference/api-alb-2020-06-16-associateadditionalcertificateswithlistener
		if !certificateIsAlreadyAssociated {
			if err := d.waitForListenerReady(ctx, cloudListenerId); err != nil {
				return err
			}

			associateAdditionalCertificatesFromListenerReq := &alialb.AssociateAdditionalCertificatesWithListenerRequest{
				ListenerId: tea.String(cloudListenerId),
				Certificates: []*alialb.AssociateAdditionalCertificatesWithListenerRequestCertificates{
					{
						CertificateId: tea.String(cloudCertId),
					},
				},
			}
			associateAdditionalCertificatesFromListenerResp, err := d.sdkClients.ALB.AssociateAdditionalCertificatesWithListenerWithContext(ctx, associateAdditionalCertificatesFromListenerReq, &dara.RuntimeOptions{})
			d.logger.Debug("sdk request 'alb.AssociateAdditionalCertificatesWithListener'", slog.Any("request", associateAdditionalCertificatesFromListenerReq), slog.Any("response", associateAdditionalCertificatesFromListenerResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'alb.AssociateAdditionalCertificatesWithListener': %w", err)
			}
		}

		// 解除关联监听和扩展证书
		// REF: https://help.aliyun.com/zh/slb/application-load-balancer/developer-reference/api-alb-2020-06-16-dissociateadditionalcertificatesfromlistener
		if !certificateIsAlreadyAssociated && len(certificateIdsToDissociate) > 0 {
			d.logger.Info("found listener certificates to dissociate", slog.Any("certificateIds", certificateIdsToDissociate))

			const MAX_CERT_PER_REQUEST = 10
			certIdChunks := lo.Chunk(certificateIdsToDissociate, MAX_CERT_PER_REQUEST)
			for _, certIds := range certIdChunks {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					if err := d.waitForListenerReady(ctx, cloudListenerId); err != nil {
						return err
					}

					dissociateAdditionalCertificatesFromListenerReq := &alialb.DissociateAdditionalCertificatesFromListenerRequest{
						ListenerId: tea.String(cloudListenerId),
						Certificates: lo.Map(certIds, func(certId string, _ int) *alialb.DissociateAdditionalCertificatesFromListenerRequestCertificates {
							return &alialb.DissociateAdditionalCertificatesFromListenerRequestCertificates{
								CertificateId: tea.String(certId),
							}
						}),
					}
					dissociateAdditionalCertificatesFromListenerResp, err := d.sdkClients.ALB.DissociateAdditionalCertificatesFromListenerWithContext(ctx, dissociateAdditionalCertificatesFromListenerReq, &dara.RuntimeOptions{})
					d.logger.Debug("sdk request 'alb.DissociateAdditionalCertificatesFromListener'", slog.Any("request", dissociateAdditionalCertificatesFromListenerReq), slog.Any("response", dissociateAdditionalCertificatesFromListenerResp))
					if err != nil {
						return fmt.Errorf("failed to execute sdk request 'alb.DissociateAdditionalCertificatesFromListener': %w", err)
					}
				}
			}
		}
	}

	return nil
}

func (d *Deployer) waitForListenerReady(ctx context.Context, cloudListenerId string) error {
	// 查询监听的属性，直到监听状态不再为 "Configuring"
	// REF: https://help.aliyun.com/zh/slb/application-load-balancer/developer-reference/api-alb-2020-06-16-getlistenerattribute
	if _, err := xwait.UntilWithContext(ctx, func(_ context.Context, _ int) (bool, error) {
		getListenerAttributeReq := &alialb.GetListenerAttributeRequest{
			ListenerId: tea.String(cloudListenerId),
		}
		getListenerAttributeResp, err := d.sdkClients.ALB.GetListenerAttributeWithContext(ctx, getListenerAttributeReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'alb.GetListenerAttribute'", slog.Any("request", getListenerAttributeReq), slog.Any("response", getListenerAttributeResp))
		if err != nil {
			return false, fmt.Errorf("failed to execute sdk request 'alb.GetListenerAttribute': %w", err)
		}

		if tea.StringValue(getListenerAttributeResp.Body.ListenerStatus) != "Configuring" {
			return true, nil
		}

		d.logger.Info("waiting for aliyun alb listener's status to not be 'Configuring' ...")
		return false, nil
	}, 10*time.Second); err != nil {
		return err
	}

	return nil
}

func createSDKClients(accessKeyId, accessKeySecret, region string) (*wSDKClients, error) {
	wsdk := &wSDKClients{}

	{
		// 接入点一览 https://api.aliyun.com/product/Alb
		var endpoint string
		switch region {
		case "", "cn-hangzhou-finance":
			endpoint = "alb.cn-hangzhou.aliyuncs.com"
		default:
			endpoint = fmt.Sprintf("alb.%s.aliyuncs.com", region)
		}

		config := &aliopen.Config{
			AccessKeyId:     tea.String(accessKeyId),
			AccessKeySecret: tea.String(accessKeySecret),
			Endpoint:        tea.String(endpoint),
		}

		client, err := alialb.NewClient(config)
		if err != nil {
			return nil, err
		}

		wsdk.ALB = client
	}

	{
		// 接入点一览 https://api.aliyun.com/product/cas
		var endpoint string
		if !strings.HasPrefix(region, "cn-") {
			endpoint = "cas.ap-southeast-1.aliyuncs.com"
		} else {
			endpoint = "cas.aliyuncs.com"
		}

		config := &aliopen.Config{
			Endpoint:        tea.String(endpoint),
			AccessKeyId:     tea.String(accessKeyId),
			AccessKeySecret: tea.String(accessKeySecret),
		}

		client, err := alicas.NewClient(config)
		if err != nil {
			return nil, err
		}

		wsdk.CAS = client
	}

	return wsdk, nil
}
