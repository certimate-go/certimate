package huaweicloudelb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	hcelb "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3"
	hcelbmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3/model"
	hcelbregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3/region"
	hciam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	hciammodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	hciamregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/huaweicloud-elb"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/huaweicloud-elb/internal"
)

type DeployerConfig struct {
	// 华为云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 华为云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 华为云企业项目 ID。
	EnterpriseProjectId string `json:"enterpriseProjectId,omitempty"`
	// 华为云区域。
	Region string `json:"region"`
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// 证书 ID。
	// 部署资源类型为 [RESOURCE_TYPE_CERTIFICATE] 时必填。
	CertificateId string `json:"certificateId,omitempty"`
	// 负载均衡器 ID。
	// 部署资源类型为 [RESOURCE_TYPE_LOADBALANCER] 时必填。
	LoadbalancerId string `json:"loadbalancerId,omitempty"`
	// 负载均衡监听 ID。
	// 部署资源类型为 [RESOURCE_TYPE_LISTENER] 时必填。
	ListenerId string `json:"listenerId,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *internal.ElbClient
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
		AccessKeyId:         config.AccessKeyId,
		SecretAccessKey:     config.SecretAccessKey,
		EnterpriseProjectId: config.EnterpriseProjectId,
		Region:              config.Region,
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
	// 根据部署资源类型决定部署方式
	switch d.config.ResourceType {
	case RESOURCE_TYPE_CERTIFICATE:
		if err := d.deployToCertificate(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	case RESOURCE_TYPE_LOADBALANCER:
		if err := d.deployToLoadbalancer(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	case RESOURCE_TYPE_LISTENER:
		if err := d.deployToListener(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported resource type '%s'", d.config.ResourceType)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) deployToLoadbalancer(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.LoadbalancerId == "" {
		return errors.New("config `loadbalancerId` is required")
	}

	// 查询负载均衡器详情
	// REF: https://support.huaweicloud.com/api-elb/ShowLoadBalancer.html
	showLoadBalancerReq := &hcelbmodel.ShowLoadBalancerRequest{
		LoadbalancerId: d.config.LoadbalancerId,
	}
	showLoadBalancerResp, err := d.sdkClient.ShowLoadBalancer(showLoadBalancerReq)
	d.logger.Debug("sdk request 'elb.ShowLoadBalancer'", slog.Any("request", showLoadBalancerReq), slog.Any("response", showLoadBalancerResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'elb.ShowLoadBalancer': %w", err)
	}

	// 查询监听器列表
	// REF: https://support.huaweicloud.com/api-elb/ListListeners.html
	listenerIds := make([]string, 0)
	listListenersMarker := (*string)(nil)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		listListenersReq := &hcelbmodel.ListListenersRequest{
			Marker:         listListenersMarker,
			Limit:          lo.ToPtr(int32(2000)),
			Protocol:       &[]string{"HTTPS", "TERMINATED_HTTPS"},
			LoadbalancerId: &[]string{showLoadBalancerResp.Loadbalancer.Id},
		}
		if d.config.EnterpriseProjectId != "" {
			listListenersReq.EnterpriseProjectId = lo.ToPtr([]string{d.config.EnterpriseProjectId})
		}
		listListenersResp, err := d.sdkClient.ListListeners(listListenersReq)
		d.logger.Debug("sdk request 'elb.ListListeners'", slog.Any("request", listListenersReq), slog.Any("response", listListenersResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'elb.ListListeners': %w", err)
		}

		if listListenersResp.Listeners == nil {
			break
		}

		for _, listener := range *listListenersResp.Listeners {
			listenerIds = append(listenerIds, listener.Id)
		}

		if len(*listListenersResp.Listeners) == 0 || listListenersResp.PageInfo.NextMarker == nil {
			break
		}

		listListenersMarker = listListenersResp.PageInfo.NextMarker
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 遍历更新监听器证书
	if len(listenerIds) == 0 {
		d.logger.Info("no listeners to deploy")
	} else {
		d.logger.Info("found https listeners to deploy", slog.Any("listenerIds", listenerIds))
		var errs []error

		for _, listenerId := range listenerIds {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := d.updateListenerCertificate(ctx, listenerId, upres.CertId); err != nil {
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

func (d *Deployer) deployToListener(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.ListenerId == "" {
		return errors.New("config `listenerId` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 更新监听器证书
	if err := d.updateListenerCertificate(ctx, d.config.ListenerId, upres.CertId); err != nil {
		return err
	}

	return nil
}

func (d *Deployer) deployToCertificate(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.CertificateId == "" {
		return errors.New("config `certificateId` is required")
	}

	// 替换证书
	opres, err := d.sdkCertmgr.Replace(ctx, d.config.CertificateId, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to replace certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate replaced", slog.Any("result", opres))
	}

	return nil
}

func (d *Deployer) updateListenerCertificate(ctx context.Context, cloudListenerId string, cloudCertId string) error {
	// 查询监听器详情
	// REF: https://support.huaweicloud.com/api-elb/ShowListener.html
	showListenerReq := &hcelbmodel.ShowListenerRequest{
		ListenerId: cloudListenerId,
	}
	showListenerResp, err := d.sdkClient.ShowListener(showListenerReq)
	d.logger.Debug("sdk request 'elb.ShowListener'", slog.Any("request", showListenerReq), slog.Any("response", showListenerResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'elb.ShowListener': %w", err)
	}

	// 更新监听器
	// REF: https://support.huaweicloud.com/api-elb/UpdateListener.html
	updateListenerReq := &hcelbmodel.UpdateListenerRequest{
		ListenerId: cloudListenerId,
		Body: &hcelbmodel.UpdateListenerRequestBody{
			Listener: &hcelbmodel.UpdateListenerOption{
				DefaultTlsContainerRef: lo.ToPtr(cloudCertId),
			},
		},
	}
	if showListenerResp.Listener.SniContainerRefs != nil {
		if len(showListenerResp.Listener.SniContainerRefs) > 0 {
			// 如果开启 SNI，需替换同 SAN 的证书
			sniCertIds := make([]string, 0)
			sniCertIds = append(sniCertIds, cloudCertId)

			listOldCertificateReq := &hcelbmodel.ListCertificatesRequest{
				Id: &showListenerResp.Listener.SniContainerRefs,
			}
			listOldCertificateResp, err := d.sdkClient.ListCertificates(listOldCertificateReq)
			d.logger.Debug("sdk request 'elb.ListCertificates'", slog.Any("request", listOldCertificateReq), slog.Any("response", listOldCertificateResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'elb.ListCertificates': %w", err)
			}

			showNewCertificateReq := &hcelbmodel.ShowCertificateRequest{
				CertificateId: cloudCertId,
			}
			showNewCertificateResp, err := d.sdkClient.ShowCertificate(showNewCertificateReq)
			d.logger.Debug("sdk request 'elb.ShowCertificate'", slog.Any("request", showNewCertificateReq), slog.Any("response", showNewCertificateResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'elb.ShowCertificate': %w", err)
			}

			for _, oldCertInfo := range *listOldCertificateResp.Certificates {
				newCertInfo := showNewCertificateResp.Certificate

				if oldCertInfo.SubjectAlternativeNames != nil && newCertInfo.SubjectAlternativeNames != nil {
					if strings.Join(*oldCertInfo.SubjectAlternativeNames, ",") == strings.Join(*newCertInfo.SubjectAlternativeNames, ",") {
						continue
					}
				} else {
					if oldCertInfo.Domain == newCertInfo.Domain {
						continue
					}
				}

				sniCertIds = append(sniCertIds, oldCertInfo.Id)
			}

			updateListenerReq.Body.Listener.SniContainerRefs = &sniCertIds
		}

		if showListenerResp.Listener.SniMatchAlgo != "" {
			updateListenerReq.Body.Listener.SniMatchAlgo = lo.ToPtr(showListenerResp.Listener.SniMatchAlgo)
		}
	}
	updateListenerResp, err := d.sdkClient.UpdateListener(updateListenerReq)
	d.logger.Debug("sdk request 'elb.UpdateListener'", slog.Any("request", updateListenerReq), slog.Any("response", updateListenerResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'elb.UpdateListener': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*internal.ElbClient, error) {
	projectId, err := getSDKProjectId(accessKeyId, secretAccessKey, region)
	if err != nil {
		return nil, err
	}

	auth, err := basic.NewCredentialsBuilder().
		WithAk(accessKeyId).
		WithSk(secretAccessKey).
		WithProjectId(projectId).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	hcRegion, err := hcelbregion.SafeValueOf(region)
	if err != nil {
		return nil, err
	}

	hcClient, err := hcelb.ElbClientBuilder().
		WithRegion(hcRegion).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	client := internal.NewElbClient(hcClient)
	return client, nil
}

func getSDKProjectId(accessKeyId, secretAccessKey, region string) (string, error) {
	if region == "" {
		region = "cn-north-4" // IAM 服务默认区域：华北四北京
	}

	auth, err := global.NewCredentialsBuilder().
		WithAk(accessKeyId).
		WithSk(secretAccessKey).
		SafeBuild()
	if err != nil {
		return "", err
	}

	hcRegion, err := hciamregion.SafeValueOf(region)
	if err != nil {
		return "", err
	}

	hcClient, err := hciam.IamClientBuilder().
		WithRegion(hcRegion).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return "", err
	}

	client := hciam.NewIamClient(hcClient)

	request := &hciammodel.KeystoneListProjectsRequest{
		Name: &region,
	}
	response, err := client.KeystoneListProjects(request)
	if err != nil {
		return "", err
	} else if response.Projects == nil || len(*response.Projects) == 0 {
		return "", errors.New("huaweicloud: no project found")
	}

	return (*response.Projects)[0].Id, nil
}
