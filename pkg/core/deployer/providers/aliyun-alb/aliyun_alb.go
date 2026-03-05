package aliyunalb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	alialb "github.com/alibabacloud-go/alb-20200616/v2/client"
	alicas "github.com/alibabacloud-go/cas-20200407/v4/client"
	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aliyun-cas"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-alb/internal"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
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
	// 部署资源类型为 [RESOURCE_TYPE_LOADBALANCER] 时必填。
	LoadbalancerId string `json:"loadbalancerId,omitempty"`
	// 负载均衡监听 ID。
	// 部署资源类型为 [RESOURCE_TYPE_LISTENER] 时必填。
	ListenerId string `json:"listenerId,omitempty"`
	// SNI 域名（支持泛域名）。
	// 部署资源类型为 [RESOURCE_TYPE_LOADBALANCER]、[RESOURCE_TYPE_LISTENER] 时选填。
	Domain string `json:"domain,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClients *wSDKClients
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

type wSDKClients struct {
	ALB *internal.AlbClient
	CAS *internal.CasClient
}

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	clients, err := createSDKClients(config.AccessKeyId, config.AccessKeySecret, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
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

	// 根据部署资源类型决定部署方式
	switch d.config.ResourceType {
	case RESOURCE_TYPE_LOADBALANCER:
		if err := d.deployToLoadbalancer(ctx, upres.ExtendedData["CertIdentifier"].(string), certX509.DNSNames); err != nil {
			return nil, err
		}

	case RESOURCE_TYPE_LISTENER:
		if err := d.deployToListener(ctx, upres.ExtendedData["CertIdentifier"].(string), certX509.DNSNames); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported resource type '%s'", d.config.ResourceType)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) deployToLoadbalancer(ctx context.Context, cloudCertId string, cloudCertSANs []string) error {
	if d.config.LoadbalancerId == "" {
		return errors.New("config `loadbalancerId` is required")
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
		return errors.New("config `listenerId` is required")
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
		listenerCertificates := make([]alialb.ListListenerCertificatesResponseBodyCertificates, 0)
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

			for _, listenerCertificate := range listListenerCertificatesResp.Body.Certificates {
				listenerCertificates = append(listenerCertificates, *listenerCertificate)
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
		if len(listenerCertificates) > 0 {
			d.logger.Info("found listener certificates to deploy", slog.Any("listenerCertificates", listenerCertificates))
			var errs []error

			for _, listenerCertificate := range listenerCertificates {
				if tea.BoolValue(listenerCertificate.IsDefault) {
					continue
				}

				if !strings.EqualFold(tea.StringValue(listenerCertificate.Status), "Associated") {
					continue
				}

				if tea.StringValue(listenerCertificate.CertificateId) == cloudCertId {
					certificateIsAlreadyAssociated = true
					break
				}

				certificateId := strings.SplitN(tea.StringValue(listenerCertificate.CertificateId), "-", 2)[0]
				certificateIdAsInt64, err := strconv.ParseInt(certificateId, 10, 64)
				if err != nil {
					errs = append(errs, err)
					continue
				}

				getCertificateDetailReq := &alicas.GetCertificateDetailRequest{
					CertificateId: tea.Int64(certificateIdAsInt64),
				}
				getCertificateDetailResp, err := d.sdkClients.CAS.GetCertificateDetailWithContext(ctx, getCertificateDetailReq, &dara.RuntimeOptions{})
				d.logger.Debug("sdk request 'cas.GetCertificateDetail'", slog.Any("request", getCertificateDetailReq), slog.Any("response", getCertificateDetailResp))
				if err != nil {
					if sdkerr, ok := err.(*tea.SDKError); ok {
						if tea.IntValue(sdkerr.StatusCode) == 400 && tea.StringValue(sdkerr.Code) == "NotFound" {
							continue
						}
					}

					errs = append(errs, fmt.Errorf("failed to execute sdk request 'cas.GetCertificateDetail': %w", err))
					continue
				} else {
					certCNMatched := tea.StringValue(getCertificateDetailResp.Body.CommonName) == d.config.Domain
					certSANDiff, _ := lo.Difference(tea.StringSliceValue(getCertificateDetailResp.Body.SubjectAlternativeNames), cloudCertSANs)
					if certCNMatched || len(certSANDiff) == 0 {
						certificateIdsToDissociate = append(certificateIdsToDissociate, certificateId)
						continue
					}

					certNotAfter := time.Unix(tea.Int64Value(getCertificateDetailResp.Body.NotAfter)/1000, 0)
					if certNotAfter.Before(time.Now()) {
						certificateIdsToDissociate = append(certificateIdsToDissociate, certificateId)
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
			if err := d.waitForListenerReady(ctx, cloudListenerId); err != nil {
				return err
			}

			dissociateAdditionalCertificatesFromListenerReq := &alialb.DissociateAdditionalCertificatesFromListenerRequest{
				ListenerId: tea.String(cloudListenerId),
				Certificates: lo.Map(certificateIdsToDissociate, func(certificateId string, _ int) *alialb.DissociateAdditionalCertificatesFromListenerRequestCertificates {
					return &alialb.DissociateAdditionalCertificatesFromListenerRequestCertificates{
						CertificateId: tea.String(certificateId),
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
	}, time.Second*5); err != nil {
		return err
	}

	return nil
}

func createSDKClients(accessKeyId, accessKeySecret, region string) (*wSDKClients, error) {
	// 接入点一览 https://api.aliyun.com/product/Alb
	var albEndpoint string
	switch region {
	case "", "cn-hangzhou-finance":
		albEndpoint = "alb.cn-hangzhou.aliyuncs.com"
	default:
		albEndpoint = fmt.Sprintf("alb.%s.aliyuncs.com", region)
	}

	albConfig := &aliopen.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(albEndpoint),
	}
	albClient, err := internal.NewAlbClient(albConfig)
	if err != nil {
		return nil, err
	}

	// 接入点一览 https://api.aliyun.com/product/cas
	var casEndpoint string
	if !strings.HasPrefix(region, "cn-") {
		casEndpoint = "cas.ap-southeast-1.aliyuncs.com"
	} else {
		casEndpoint = "cas.aliyuncs.com"
	}

	casConfig := &aliopen.Config{
		Endpoint:        tea.String(casEndpoint),
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
	}
	casClient, err := internal.NewCasClient(casConfig)
	if err != nil {
		return nil, err
	}

	return &wSDKClients{
		ALB: albClient,
		CAS: casClient,
	}, nil
}
