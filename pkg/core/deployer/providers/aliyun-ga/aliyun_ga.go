package aliyunga

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/samber/lo"

	aliga "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/alibabacloud-go/ga-20191120/v4/client"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aliyun-cas"
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
	// 部署目标。
	DeployTarget string `json:"deployTarget"`
	// 全球加速实例 ID。
	AcceleratorId string `json:"acceleratorId"`
	// 全球加速监听 ID。
	// 部署目标为 [DEPLOY_TARGET_LISTENER] 时必填。
	ListenerId string `json:"listenerId,omitempty"`
	// SNI 域名（不支持泛域名）。
	// 部署目标为 [DEPLOY_TARGET_ACCELERATOR]、[DEPLOY_TARGET_LISTENER] 时选填。
	Domain string `json:"domain,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *aliga.Client
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
		ResourceGroupId: config.ResourceGroupId,
		Region:          "cn-hangzhou",
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
	case DEPLOY_TARGET_ACCELERATOR:
		if err := d.deployToAccelerator(ctx, upres.ExtendedData["CertIdWithRegion"].(string)); err != nil {
			return nil, err
		}

	case DEPLOY_TARGET_LISTENER:
		if err := d.deployToListener(ctx, upres.ExtendedData["CertIdWithRegion"].(string)); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported deploy target '%s'", d.config.DeployTarget)
	}

	return &DeployResult{}, nil
}

func (d *Deployer) deployToAccelerator(ctx context.Context, cloudCertId string) error {
	if d.config.AcceleratorId == "" {
		return fmt.Errorf("config `acceleratorId` is required")
	}

	// 查询 HTTPS 监听列表
	// REF: https://help.aliyun.com/zh/ga/developer-reference/api-ga-2019-11-20-listlisteners
	listenerIds := make([]string, 0)
	listListenersPageNumber := 1
	listListenersPageSize := 50
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		listListenersReq := &aliga.ListListenersRequest{
			RegionId:      tea.String("cn-hangzhou"),
			AcceleratorId: tea.String(d.config.AcceleratorId),
			PageNumber:    tea.Int32(int32(listListenersPageNumber)),
			PageSize:      tea.Int32(int32(listListenersPageSize)),
		}
		listListenersResp, err := d.sdkClient.ListListenersWithContext(ctx, listListenersReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'ga.ListListeners'", slog.Any("request", listListenersReq), slog.Any("response", listListenersResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'ga.ListListeners': %w", err)
		}

		if listListenersResp.Body == nil {
			break
		}

		for _, listener := range listListenersResp.Body.Listeners {
			if strings.EqualFold(tea.StringValue(listener.Protocol), "https") {
				listenerIds = append(listenerIds, tea.StringValue(listener.ListenerId))
			}
		}

		if len(listListenersResp.Body.Listeners) < listListenersPageSize {
			break
		}

		listListenersPageNumber++
	}

	// 批量更新监听证书
	if len(listenerIds) == 0 {
		d.logger.Info("no ga listeners to deploy")
	} else {
		d.logger.Info("found ga listeners to deploy", slog.Any("listenerIds", listenerIds))

		if err := xloop.ForRangeAllWithContext(ctx, listenerIds, func(ctx context.Context, listenerId string, _ int) error {
			return d.updateListenerCertificate(ctx, d.config.AcceleratorId, listenerId, cloudCertId)
		}); err != nil {
			return err
		}
	}

	return nil
}

func (d *Deployer) deployToListener(ctx context.Context, cloudCertId string) error {
	if d.config.AcceleratorId == "" {
		return fmt.Errorf("config `acceleratorId` is required")
	}
	if d.config.ListenerId == "" {
		return fmt.Errorf("config `listenerId` is required")
	}

	// 更新监听
	if err := d.updateListenerCertificate(ctx, d.config.AcceleratorId, d.config.ListenerId, cloudCertId); err != nil {
		return err
	}

	return nil
}

func (d *Deployer) updateListenerCertificate(ctx context.Context, cloudAcceleratorId string, cloudListenerId string, cloudCertId string) error {
	// 查询监听绑定的证书列表
	// REF: https://help.aliyun.com/zh/ga/developer-reference/api-ga-2019-11-20-listlistenercertificates
	listenerDefaultCertificate := (*aliga.ListListenerCertificatesResponseBodyCertificates)(nil)
	listenerAdditionalCertificates := make([]*aliga.ListListenerCertificatesResponseBodyCertificates, 0)
	listListenerCertificatesNextToken := (*string)(nil)
	for {
		listListenerCertificatesReq := &aliga.ListListenerCertificatesRequest{
			RegionId:      tea.String("cn-hangzhou"),
			AcceleratorId: tea.String(cloudAcceleratorId),
			ListenerId:    tea.String(cloudListenerId),
			NextToken:     listListenerCertificatesNextToken,
			MaxResults:    tea.Int32(20),
		}
		listListenerCertificatesResp, err := d.sdkClient.ListListenerCertificatesWithContext(ctx, listListenerCertificatesReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'ga.ListListenerCertificates'", slog.Any("request", listListenerCertificatesReq), slog.Any("response", listListenerCertificatesResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'ga.ListListenerCertificates': %w", err)
		}

		if listListenerCertificatesResp.Body == nil {
			break
		}

		for _, certItem := range listListenerCertificatesResp.Body.Certificates {
			if tea.BoolValue(certItem.IsDefault) {
				listenerDefaultCertificate = certItem
			} else {
				listenerAdditionalCertificates = append(listenerAdditionalCertificates, certItem)
			}
		}

		if len(listListenerCertificatesResp.Body.Certificates) == 0 || listListenerCertificatesResp.Body.NextToken == nil {
			break
		}

		listListenerCertificatesNextToken = listListenerCertificatesResp.Body.NextToken
	}

	if d.config.Domain == "" {
		// 未指定 SNI，只需部署到监听器
		if listenerDefaultCertificate != nil && tea.StringValue(listenerDefaultCertificate.CertificateId) == cloudCertId {
			d.logger.Info("no need to deploy ga listener default certificate")
			return nil
		}
		return d.updateListenerDefaultCertificate(ctx, cloudListenerId, cloudCertId)
	} else {
		// 指定 SNI，需部署到扩展域名
		if lo.SomeBy(listenerAdditionalCertificates, func(item *aliga.ListListenerCertificatesResponseBodyCertificates) bool {
			return tea.StringValue(item.CertificateId) == cloudCertId
		}) {
			d.logger.Info("no need to deploy ga listener sni certificate")
			return nil
		}

		added := lo.SomeBy(listenerAdditionalCertificates, func(item *aliga.ListListenerCertificatesResponseBodyCertificates) bool {
			return tea.StringValue(item.Domain) == d.config.Domain
		})
		return d.updateListenerSniCertificate(ctx, cloudAcceleratorId, cloudListenerId, cloudCertId, added)
	}
}

func (d *Deployer) updateListenerDefaultCertificate(ctx context.Context, cloudListenerId string, cloudCertId string) error {
	// 修改监听的属性
	// REF: https://help.aliyun.com/zh/ga/developer-reference/api-ga-2019-11-20-updatelistener
	updateListenerReq := &aliga.UpdateListenerRequest{
		RegionId:   tea.String("cn-hangzhou"),
		ListenerId: tea.String(cloudListenerId),
		Certificates: []*aliga.UpdateListenerRequestCertificates{{
			Id: tea.String(cloudCertId),
		}},
	}
	updateListenerResp, err := d.sdkClient.UpdateListenerWithContext(ctx, updateListenerReq, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'ga.UpdateListener'", slog.Any("request", updateListenerReq), slog.Any("response", updateListenerResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'ga.UpdateListener': %w", err)
	}

	return nil
}

func (d *Deployer) updateListenerSniCertificate(ctx context.Context, cloudAcceleratorId string, cloudListenerId string, cloudCertId string, added bool) error {
	if added {
		// 为监听替换扩展证书
		// REF: https://help.aliyun.com/zh/ga/developer-reference/api-ga-2019-11-20-updateadditionalcertificatewithlistener
		updateAdditionalCertificateWithListenerReq := &aliga.UpdateAdditionalCertificateWithListenerRequest{
			RegionId:      tea.String("cn-hangzhou"),
			AcceleratorId: tea.String(cloudAcceleratorId),
			ListenerId:    tea.String(cloudListenerId),
			CertificateId: tea.String(cloudCertId),
			Domain:        tea.String(d.config.Domain),
		}
		updateAdditionalCertificateWithListenerResp, err := d.sdkClient.UpdateAdditionalCertificateWithListenerWithContext(ctx, updateAdditionalCertificateWithListenerReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'ga.UpdateAdditionalCertificateWithListener'", slog.Any("request", updateAdditionalCertificateWithListenerReq), slog.Any("response", updateAdditionalCertificateWithListenerResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'ga.UpdateAdditionalCertificateWithListener': %w", err)
		}
	} else {
		// 为监听绑定扩展证书
		// REF: https://help.aliyun.com/zh/ga/developer-reference/api-ga-2019-11-20-associateadditionalcertificateswithlistener
		associateAdditionalCertificatesWithListenerReq := &aliga.AssociateAdditionalCertificatesWithListenerRequest{
			RegionId:      tea.String("cn-hangzhou"),
			AcceleratorId: tea.String(cloudAcceleratorId),
			ListenerId:    tea.String(cloudListenerId),
			Certificates: []*aliga.AssociateAdditionalCertificatesWithListenerRequestCertificates{{
				Id:     tea.String(cloudCertId),
				Domain: tea.String(d.config.Domain),
			}},
		}
		associateAdditionalCertificatesWithListenerResp, err := d.sdkClient.AssociateAdditionalCertificatesWithListenerWithContext(ctx, associateAdditionalCertificatesWithListenerReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'ga.AssociateAdditionalCertificatesWithListener'", slog.Any("request", associateAdditionalCertificatesWithListenerReq), slog.Any("response", associateAdditionalCertificatesWithListenerResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'ga.AssociateAdditionalCertificatesWithListener': %w", err)
		}
	}

	return nil
}

func createSDKClient(accessKeyId, accessKeySecret string) (*aliga.Client, error) {
	// 接入点一览 https://api.aliyun.com/product/Ga
	config := &aliopen.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String("ga.cn-hangzhou.aliyuncs.com"),
	}

	client, err := aliga.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
