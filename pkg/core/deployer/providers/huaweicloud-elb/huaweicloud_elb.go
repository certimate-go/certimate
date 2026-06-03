package huaweicloudelb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	hwelbmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3/model"
	hwelbregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3/region"
	hwiam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	hwiammodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	hwiamregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"
	"github.com/samber/lo"

	hwelb "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/huaweicloud-elb"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
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
	// 部署目标。
	DeployTarget string `json:"deployTarget"`
	// 证书 ID。
	// 部署目标为 [DEPLOY_TARGET_CERTIFICATE] 时必填。
	CertificateId string `json:"certificateId,omitempty"`
	// 负载均衡器 ID。
	// 部署目标为 [DEPLOY_TARGET_LOADBALANCER] 时必填。
	LoadbalancerId string `json:"loadbalancerId,omitempty"`
	// 负载均衡监听 ID。
	// 部署目标为 [DEPLOY_TARGET_LISTENER] 时必填。
	ListenerId string `json:"listenerId,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *hwelb.ElbClient
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*DeployResult, error) {
	// 根据部署目标决定业务流程
	switch d.config.DeployTarget {
	case DEPLOY_TARGET_CERTIFICATE:
		if err := d.deployToCertificate(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	case DEPLOY_TARGET_LOADBALANCER:
		if err := d.deployToLoadbalancer(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	case DEPLOY_TARGET_LISTENER:
		if err := d.deployToListener(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported deploy target '%s'", d.config.DeployTarget)
	}

	return &DeployResult{}, nil
}

func (d *Deployer) deployToLoadbalancer(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.LoadbalancerId == "" {
		return fmt.Errorf("config `loadbalancerId` is required")
	}

	// 查询负载均衡器详情
	// REF: https://support.huaweicloud.com/api-elb/ShowLoadBalancer.html
	showLoadBalancerReq := &hwelbmodel.ShowLoadBalancerRequest{
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

		listListenersReq := &hwelbmodel.ListListenersRequest{
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
		return fmt.Errorf("config `listenerId` is required")
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
		return fmt.Errorf("config `certificateId` is required")
	}

	// 替换证书
	rplres, err := d.sdkCertmgr.Replace(ctx, d.config.CertificateId, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to replace certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate replaced", slog.Any("result", rplres))
	}

	return nil
}

func (d *Deployer) updateListenerCertificate(ctx context.Context, cloudListenerId string, cloudCertId string) error {
	// 查询监听器详情
	// REF: https://support.huaweicloud.com/api-elb/ShowListener.html
	showListenerReq := &hwelbmodel.ShowListenerRequest{
		ListenerId: cloudListenerId,
	}
	showListenerResp, err := d.sdkClient.ShowListener(showListenerReq)
	d.logger.Debug("sdk request 'elb.ShowListener'", slog.Any("request", showListenerReq), slog.Any("response", showListenerResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'elb.ShowListener': %w", err)
	}

	// 更新监听器
	// REF: https://support.huaweicloud.com/api-elb/UpdateListener.html
	updateListenerReq := &hwelbmodel.UpdateListenerRequest{
		ListenerId: cloudListenerId,
		Body: &hwelbmodel.UpdateListenerRequestBody{
			Listener: &hwelbmodel.UpdateListenerOption{
				DefaultTlsContainerRef: lo.ToPtr(cloudCertId),
			},
		},
	}
	if showListenerResp.Listener.SniContainerRefs != nil {
		if len(showListenerResp.Listener.SniContainerRefs) > 0 {
			// 如果开启 SNI，需替换同 SAN 的证书
			sniCertIds := make([]string, 0)
			sniCertIds = append(sniCertIds, cloudCertId)

			listOldCertificateReq := &hwelbmodel.ListCertificatesRequest{
				Id: &showListenerResp.Listener.SniContainerRefs,
			}
			listOldCertificateResp, err := d.sdkClient.ListCertificates(listOldCertificateReq)
			d.logger.Debug("sdk request 'elb.ListCertificates'", slog.Any("request", listOldCertificateReq), slog.Any("response", listOldCertificateResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'elb.ListCertificates': %w", err)
			}

			showNewCertificateReq := &hwelbmodel.ShowCertificateRequest{
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

func createSDKClient(accessKeyId, secretAccessKey, region string) (*hwelb.ElbClient, error) {
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

	hcRegion, err := hwelbregion.SafeValueOf(region)
	if err != nil {
		return nil, err
	}

	hcClient, err := hwelb.ElbClientBuilder().
		WithRegion(hcRegion).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	client := hwelb.NewElbClient(hcClient)
	return client, nil
}

func getSDKProjectId(accessKeyId, secretAccessKey, region string) (string, error) {
	if region == "" {
		region = "cn-north-4" // IAM 服务默认区域：华北北京四
	}

	auth, err := global.NewCredentialsBuilder().
		WithAk(accessKeyId).
		WithSk(secretAccessKey).
		SafeBuild()
	if err != nil {
		return "", err
	}

	hcRegion, err := hwiamregion.SafeValueOf(region)
	if err != nil {
		return "", err
	}

	hcClient, err := hwiam.IamClientBuilder().
		WithRegion(hcRegion).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return "", err
	}

	client := hwiam.NewIamClient(hcClient)

	request := &hwiammodel.KeystoneListProjectsRequest{
		Name: &region,
	}
	response, err := client.KeystoneListProjects(request)
	if err != nil {
		return "", err
	} else if response.Projects == nil || len(*response.Projects) == 0 {
		return "", fmt.Errorf("huaweicloud: no project found")
	}

	return (*response.Projects)[0].Id, nil
}
