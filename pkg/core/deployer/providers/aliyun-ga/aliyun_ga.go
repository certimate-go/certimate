package aliyunga

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	aliga "github.com/alibabacloud-go/ga-20191120/v4/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aliyun-cas"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-ga/internal"
)

type DeployerConfig struct {
	// 阿里云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 阿里云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 阿里云资源组 ID。
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// 全球加速实例 ID。
	AcceleratorId string `json:"acceleratorId"`
	// 全球加速监听 ID。
	// 部署资源类型为 [RESOURCE_TYPE_LISTENER] 时必填。
	ListenerId string `json:"listenerId,omitempty"`
	// SNI 域名（不支持泛域名）。
	// 部署资源类型为 [RESOURCE_TYPE_ACCELERATOR]、[RESOURCE_TYPE_LISTENER] 时选填。
	Domain string `json:"domain,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *internal.GaClient
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM, issuerCertificatePEM string) (*deployer.DeployResult, error) {
	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 根据部署资源类型决定部署方式
	switch d.config.ResourceType {
	case RESOURCE_TYPE_ACCELERATOR:
		if err := d.deployToAccelerator(ctx, upres.ExtendedData["CertIdentifier"].(string)); err != nil {
			return nil, err
		}

	case RESOURCE_TYPE_LISTENER:
		if err := d.deployToListener(ctx, upres.ExtendedData["CertIdentifier"].(string)); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported resource type '%s'", d.config.ResourceType)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) deployToAccelerator(ctx context.Context, cloudCertId string) error {
	if d.config.AcceleratorId == "" {
		return errors.New("config `acceleratorId` is required")
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
		listListenersResp, err := d.sdkClient.ListListeners(listListenersReq)
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

	// 遍历更新监听证书
	if len(listenerIds) == 0 {
		d.logger.Info("no ga listeners to deploy")
	} else {
		var errs []error
		d.logger.Info("found https listeners to deploy", slog.Any("listenerIds", listenerIds))

		for _, listenerId := range listenerIds {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := d.updateListenerCertificate(ctx, d.config.AcceleratorId, listenerId, cloudCertId); err != nil {
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
	if d.config.AcceleratorId == "" {
		return errors.New("config `acceleratorId` is required")
	}
	if d.config.ListenerId == "" {
		return errors.New("config `listenerId` is required")
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
		listListenerCertificatesResp, err := d.sdkClient.ListListenerCertificates(listListenerCertificatesReq)
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
			d.logger.Info("no need to update ga listener default certificate")
			return nil
		}

		// 修改监听的属性
		// REF: https://help.aliyun.com/zh/ga/developer-reference/api-ga-2019-11-20-updatelistener
		updateListenerReq := &aliga.UpdateListenerRequest{
			RegionId:   tea.String("cn-hangzhou"),
			ListenerId: tea.String(cloudListenerId),
			Certificates: []*aliga.UpdateListenerRequestCertificates{{
				Id: tea.String(cloudCertId),
			}},
		}
		updateListenerResp, err := d.sdkClient.UpdateListener(updateListenerReq)
		d.logger.Debug("sdk request 'ga.UpdateListener'", slog.Any("request", updateListenerReq), slog.Any("response", updateListenerResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'ga.UpdateListener': %w", err)
		}
	} else {
		// 指定 SNI，需部署到扩展域名
		if lo.SomeBy(listenerAdditionalCertificates, func(item *aliga.ListListenerCertificatesResponseBodyCertificates) bool {
			return tea.StringValue(item.CertificateId) == cloudCertId
		}) {
			d.logger.Info("no need to update ga listener additional certificate")
			return nil
		}

		if lo.SomeBy(listenerAdditionalCertificates, func(item *aliga.ListListenerCertificatesResponseBodyCertificates) bool {
			return tea.StringValue(item.Domain) == d.config.Domain
		}) {
			// 为监听替换扩展证书
			// REF: https://help.aliyun.com/zh/ga/developer-reference/api-ga-2019-11-20-updateadditionalcertificatewithlistener
			updateAdditionalCertificateWithListenerReq := &aliga.UpdateAdditionalCertificateWithListenerRequest{
				RegionId:      tea.String("cn-hangzhou"),
				AcceleratorId: tea.String(cloudAcceleratorId),
				ListenerId:    tea.String(cloudListenerId),
				CertificateId: tea.String(cloudCertId),
				Domain:        tea.String(d.config.Domain),
			}
			updateAdditionalCertificateWithListenerResp, err := d.sdkClient.UpdateAdditionalCertificateWithListener(updateAdditionalCertificateWithListenerReq)
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
			associateAdditionalCertificatesWithListenerResp, err := d.sdkClient.AssociateAdditionalCertificatesWithListener(associateAdditionalCertificatesWithListenerReq)
			d.logger.Debug("sdk request 'ga.AssociateAdditionalCertificatesWithListener'", slog.Any("request", associateAdditionalCertificatesWithListenerReq), slog.Any("response", associateAdditionalCertificatesWithListenerResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'ga.AssociateAdditionalCertificatesWithListener': %w", err)
			}
		}
	}

	return nil
}

func createSDKClient(accessKeyId, accessKeySecret string) (*internal.GaClient, error) {
	// 接入点一览 https://api.aliyun.com/product/Ga
	config := &aliopen.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String("ga.cn-hangzhou.aliyuncs.com"),
	}

	client, err := internal.NewGaClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
