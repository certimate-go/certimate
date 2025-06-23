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
	alicas "github.com/alibabacloud-go/cas-20200407/v3/client"
	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"golang.org/x/exp/slices"

	"github.com/certimate-go/certimate/pkg/core"
	sslmgrsp "github.com/certimate-go/certimate/pkg/core/ssl-manager/providers/aliyun-cas"
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
	// 部署资源类型为 [RESOURCE_TYPE_LOADBALANCER] 时必填。
	LoadbalancerId string `json:"loadbalancerId,omitempty"`
	// 负载均衡监听 ID。
	// 部署资源类型为 [RESOURCE_TYPE_LISTENER] 时必填。
	ListenerId string `json:"listenerId,omitempty"`
	// SNI 域名（支持泛域名）。
	// 部署资源类型为 [RESOURCE_TYPE_LOADBALANCER]、[RESOURCE_TYPE_LISTENER] 时选填。
	Domain string `json:"domain,omitempty"`
}

type SSLDeployerProvider struct {
	config     *SSLDeployerProviderConfig
	logger     *slog.Logger
	sdkClients *wSDKClients
	sslManager core.SSLManager
}

var _ core.SSLDeployer = (*SSLDeployerProvider)(nil)

type wSDKClients struct {
	ALB *alialb.Client
	CAS *alicas.Client
}

func NewSSLDeployerProvider(config *SSLDeployerProviderConfig) (*SSLDeployerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl deployer provider is nil")
	}

	clients, err := createSDKClients(config.AccessKeyId, config.AccessKeySecret, config.Region)
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
		sdkClients: clients,
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
	// REF: https://help.aliyun.com/zh/slb/application-load-balancer/developer-reference/api-alb-2020-06-16-getloadbalancerattribute
	getLoadBalancerAttributeReq := &alialb.GetLoadBalancerAttributeRequest{
		LoadBalancerId: tea.String(d.config.LoadbalancerId),
	}
	getLoadBalancerAttributeResp, err := d.sdkClients.ALB.GetLoadBalancerAttribute(getLoadBalancerAttributeReq)
	d.logger.Debug("sdk request 'alb.GetLoadBalancerAttribute'", slog.Any("request", getLoadBalancerAttributeReq), slog.Any("response", getLoadBalancerAttributeResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'alb.GetLoadBalancerAttribute': %w", err)
	}

	// 查询 HTTPS 监听列表
	// REF: https://help.aliyun.com/zh/slb/application-load-balancer/developer-reference/api-alb-2020-06-16-listlisteners
	listenerIds := make([]string, 0)
	listListenersLimit := int32(100)
	var listListenersToken *string = nil
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		listListenersReq := &alialb.ListListenersRequest{
			MaxResults:       tea.Int32(listListenersLimit),
			NextToken:        listListenersToken,
			LoadBalancerIds:  []*string{tea.String(d.config.LoadbalancerId)},
			ListenerProtocol: tea.String("HTTPS"),
		}
		listListenersResp, err := d.sdkClients.ALB.ListListeners(listListenersReq)
		d.logger.Debug("sdk request 'alb.ListListeners'", slog.Any("request", listListenersReq), slog.Any("response", listListenersResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'alb.ListListeners': %w", err)
		}

		if listListenersResp.Body.Listeners != nil {
			for _, listener := range listListenersResp.Body.Listeners {
				listenerIds = append(listenerIds, tea.StringValue(listener.ListenerId))
			}
		}

		if len(listListenersResp.Body.Listeners) == 0 || listListenersResp.Body.NextToken == nil {
			break
		} else {
			listListenersToken = listListenersResp.Body.NextToken
		}
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
			MaxResults:       tea.Int32(listListenersLimit),
			NextToken:        listListenersToken,
			LoadBalancerIds:  []*string{tea.String(d.config.LoadbalancerId)},
			ListenerProtocol: tea.String("QUIC"),
		}
		listListenersResp, err := d.sdkClients.ALB.ListListeners(listListenersReq)
		d.logger.Debug("sdk request 'alb.ListListeners'", slog.Any("request", listListenersReq), slog.Any("response", listListenersResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'alb.ListListeners': %w", err)
		}

		if listListenersResp.Body.Listeners != nil {
			for _, listener := range listListenersResp.Body.Listeners {
				listenerIds = append(listenerIds, tea.StringValue(listener.ListenerId))
			}
		}

		if len(listListenersResp.Body.Listeners) == 0 || listListenersResp.Body.NextToken == nil {
			break
		} else {
			listListenersToken = listListenersResp.Body.NextToken
		}
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

func (d *SSLDeployerProvider) deployToListener(ctx context.Context, cloudCertId string) error {
	if d.config.ListenerId == "" {
		return errors.New("config `listenerId` is required")
	}

	// 更新监听
	if err := d.updateListenerCertificate(ctx, d.config.ListenerId, cloudCertId); err != nil {
		return err
	}

	return nil
}

func (d *SSLDeployerProvider) updateListenerCertificate(ctx context.Context, cloudListenerId string, cloudCertId string) error {
	// 查询监听的属性
	// REF: https://help.aliyun.com/zh/slb/application-load-balancer/developer-reference/api-alb-2020-06-16-getlistenerattribute
	getListenerAttributeReq := &alialb.GetListenerAttributeRequest{
		ListenerId: tea.String(cloudListenerId),
	}
	getListenerAttributeResp, err := d.sdkClients.ALB.GetListenerAttribute(getListenerAttributeReq)
	d.logger.Debug("sdk request 'alb.GetListenerAttribute'", slog.Any("request", getListenerAttributeReq), slog.Any("response", getListenerAttributeResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'alb.GetListenerAttribute': %w", err)
	}

	if d.config.Domain == "" {
		// 未指定 SNI，只需部署到监听器

		// 修改监听的属性
		// REF: https://help.aliyun.com/zh/slb/application-load-balancer/developer-reference/api-alb-2020-06-16-updatelistenerattribute
		updateListenerAttributeReq := &alialb.UpdateListenerAttributeRequest{
			ListenerId: tea.String(cloudListenerId),
			Certificates: []*alialb.UpdateListenerAttributeRequestCertificates{{
				CertificateId: tea.String(cloudCertId),
			}},
		}
		updateListenerAttributeResp, err := d.sdkClients.ALB.UpdateListenerAttribute(updateListenerAttributeReq)
		d.logger.Debug("sdk request 'alb.UpdateListenerAttribute'", slog.Any("request", updateListenerAttributeReq), slog.Any("response", updateListenerAttributeResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'alb.UpdateListenerAttribute': %w", err)
		}
	} else {
		// 指定 SNI，需部署到扩展域名

		// 查询监听证书列表
		// REF: https://help.aliyun.com/zh/slb/application-load-balancer/developer-reference/api-alb-2020-06-16-listlistenercertificates
		listenerCertificates := make([]alialb.ListListenerCertificatesResponseBodyCertificates, 0)
		listListenerCertificatesLimit := int32(100)
		var listListenerCertificatesToken *string = nil
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			listListenerCertificatesReq := &alialb.ListListenerCertificatesRequest{
				NextToken:       listListenerCertificatesToken,
				MaxResults:      tea.Int32(listListenerCertificatesLimit),
				ListenerId:      tea.String(cloudListenerId),
				CertificateType: tea.String("Server"),
			}
			listListenerCertificatesResp, err := d.sdkClients.ALB.ListListenerCertificates(listListenerCertificatesReq)
			d.logger.Debug("sdk request 'alb.ListListenerCertificates'", slog.Any("request", listListenerCertificatesReq), slog.Any("response", listListenerCertificatesResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'alb.ListListenerCertificates': %w", err)
			}

			if listListenerCertificatesResp.Body.Certificates != nil {
				for _, listenerCertificate := range listListenerCertificatesResp.Body.Certificates {
					listenerCertificates = append(listenerCertificates, *listenerCertificate)
				}
			}

			if len(listListenerCertificatesResp.Body.Certificates) == 0 || listListenerCertificatesResp.Body.NextToken == nil {
				break
			} else {
				listListenerCertificatesToken = listListenerCertificatesResp.Body.NextToken
			}
		}

		// 遍历查询监听证书，并找出需要解除关联的证书
		// REF: https://help.aliyun.com/zh/slb/application-load-balancer/developer-reference/api-alb-2020-06-16-listlistenercertificates
		// REF: https://help.aliyun.com/zh/ssl-certificate/developer-reference/api-cas-2020-04-07-getusercertificatedetail
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

				// 监听证书 ID 格式：${证书 ID}-${地域}
				certificateId := strings.Split(tea.StringValue(listenerCertificate.CertificateId), "-")[0]
				if certificateId == cloudCertId {
					certificateIsAlreadyAssociated = true
					break
				}

				certificateIdAsInt64, err := strconv.ParseInt(certificateId, 10, 64)
				if err != nil {
					errs = append(errs, err)
					continue
				}

				getUserCertificateDetailReq := &alicas.GetUserCertificateDetailRequest{
					CertId: tea.Int64(certificateIdAsInt64),
				}
				getUserCertificateDetailResp, err := d.sdkClients.CAS.GetUserCertificateDetail(getUserCertificateDetailReq)
				d.logger.Debug("sdk request 'cas.GetUserCertificateDetail'", slog.Any("request", getUserCertificateDetailReq), slog.Any("response", getUserCertificateDetailResp))
				if err != nil {
					if sdkerr, ok := err.(*tea.SDKError); ok {
						if tea.IntValue(sdkerr.StatusCode) == 400 && tea.StringValue(sdkerr.Code) == "NotFound" {
							continue
						}
					}

					errs = append(errs, fmt.Errorf("failed to execute sdk request 'cas.GetUserCertificateDetail': %w", err))
					continue
				} else {
					certCNMatched := tea.StringValue(getUserCertificateDetailResp.Body.Common) == d.config.Domain
					certSANMatched := slices.Contains(strings.Split(tea.StringValue(getUserCertificateDetailResp.Body.Sans), ","), d.config.Domain)
					if !certCNMatched && !certSANMatched {
						continue
					}

					certEndDate, _ := time.Parse("2006-01-02", tea.StringValue(getUserCertificateDetailResp.Body.EndDate))
					if time.Now().Before(certEndDate) {
						continue
					}

					certificateIdsToDissociate = append(certificateIdsToDissociate, certificateId)
				}
			}

			if len(errs) > 0 {
				return errors.Join(errs...)
			}
		}

		// 关联监听和扩展证书
		// REF: https://help.aliyun.com/zh/slb/application-load-balancer/developer-reference/api-alb-2020-06-16-associateadditionalcertificateswithlistener
		if !certificateIsAlreadyAssociated {
			associateAdditionalCertificatesFromListenerReq := &alialb.AssociateAdditionalCertificatesWithListenerRequest{
				ListenerId: tea.String(cloudListenerId),
				Certificates: []*alialb.AssociateAdditionalCertificatesWithListenerRequestCertificates{
					{
						CertificateId: tea.String(cloudCertId),
					},
				},
			}
			associateAdditionalCertificatesFromListenerResp, err := d.sdkClients.ALB.AssociateAdditionalCertificatesWithListener(associateAdditionalCertificatesFromListenerReq)
			d.logger.Debug("sdk request 'alb.AssociateAdditionalCertificatesWithListener'", slog.Any("request", associateAdditionalCertificatesFromListenerReq), slog.Any("response", associateAdditionalCertificatesFromListenerResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'alb.AssociateAdditionalCertificatesWithListener': %w", err)
			}
		}

		// 解除关联监听和扩展证书
		// REF: https://help.aliyun.com/zh/slb/application-load-balancer/developer-reference/api-alb-2020-06-16-dissociateadditionalcertificatesfromlistener
		if !certificateIsAlreadyAssociated && len(certificateIdsToDissociate) > 0 {
			dissociateAdditionalCertificates := make([]*alialb.DissociateAdditionalCertificatesFromListenerRequestCertificates, 0)
			for _, certificateId := range certificateIdsToDissociate {
				dissociateAdditionalCertificates = append(dissociateAdditionalCertificates, &alialb.DissociateAdditionalCertificatesFromListenerRequestCertificates{
					CertificateId: tea.String(certificateId),
				})
			}

			dissociateAdditionalCertificatesFromListenerReq := &alialb.DissociateAdditionalCertificatesFromListenerRequest{
				ListenerId:   tea.String(cloudListenerId),
				Certificates: dissociateAdditionalCertificates,
			}
			dissociateAdditionalCertificatesFromListenerResp, err := d.sdkClients.ALB.DissociateAdditionalCertificatesFromListener(dissociateAdditionalCertificatesFromListenerReq)
			d.logger.Debug("sdk request 'alb.DissociateAdditionalCertificatesFromListener'", slog.Any("request", dissociateAdditionalCertificatesFromListenerReq), slog.Any("response", dissociateAdditionalCertificatesFromListenerResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'alb.DissociateAdditionalCertificatesFromListener': %w", err)
			}
		}
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
	albClient, err := alialb.NewClient(albConfig)
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
	casClient, err := alicas.NewClient(casConfig)
	if err != nil {
		return nil, err
	}

	return &wSDKClients{
		ALB: albClient,
		CAS: casClient,
	}, nil
}
