package aliyunmse

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/samber/lo"

	alimse "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/alibabacloud-go/mse-20190531/v5/client"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aliyun-cas"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
	xalibabacloud "github.com/certimate-go/certimate/pkg/utils/third-party/alibabacloud"
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
	// 服务类型。
	ServiceType string `json:"serviceType"`
	// 云原生网关 ID。
	// 服务类型为 [SERVICE_TYPE_CLOUDNATIVE] 时必填。
	GatewayId string `json:"gatewayId,omitempty"`
	// 域名匹配模式。
	// 服务类型为 [SERVICE_TYPE_CLOUDNATIVE] 时必填。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 云原生网关绑定的域名。
	// 服务类型为 [SERVICE_TYPE_CLOUDNATIVE] 时必填。
	Domain string `json:"domain,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *alimse.Client
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
		Region:          lo.Ternary(xalibabacloud.IsIntlRegion(config.Region), "ap-southeast-1", ""),
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
	switch d.config.ServiceType {
	case SERVICE_TYPE_CLOUDNATIVE:
		if err := d.deployToCloudNative(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported service type '%s'", d.config.ServiceType)
	}

	return &DeployResult{}, nil
}

func (d *Deployer) deployToCloudNative(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.GatewayId == "" {
		return fmt.Errorf("config `gatewayId` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 获取全部可部署的域名列表
	domainsInGateway, err := d.getCloudNativeAllDomainsByGatewayId(ctx, d.config.GatewayId)
	if err != nil {
		return err
	}
	if len(domainsInGateway) == 0 {
		return fmt.Errorf("could not find domains in gateway '%s'", d.config.GatewayId)
	}

	// 获取待部署的域名列表
	domains := make([]*alimse.ListGatewayDomainResponseBodyData, 0)
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if d.config.Domain == "" {
				return fmt.Errorf("config `domain` is required")
			}

			domain, ok := lo.Find(domainsInGateway, func(domain *alimse.ListGatewayDomainResponseBodyData) bool {
				return d.config.Domain == tea.StringValue(domain.Name)
			})
			if !ok {
				return fmt.Errorf("could not find domain '%s'", d.config.Domain)
			}

			domains = []*alimse.ListGatewayDomainResponseBodyData{domain}
		}

	case DOMAIN_MATCH_PATTERN_WILDCARD:
		{
			if d.config.Domain == "" {
				return fmt.Errorf("config `domain` is required")
			}

			domains = lo.Filter(domainsInGateway, func(domain *alimse.ListGatewayDomainResponseBodyData, _ int) bool {
				return xcerthostname.IsMatch(d.config.Domain, tea.StringValue(domain.Name))
			})
			if len(domains) == 0 {
				return fmt.Errorf("could not find any domains matched by wildcard")
			}
		}

	case DOMAIN_MATCH_PATTERN_CERTSAN:
		{
			domains = lo.Filter(domainsInGateway, func(domain *alimse.ListGatewayDomainResponseBodyData, _ int) bool {
				return xcerthostname.IsMatchByCertificatePEM(certPEM, tea.StringValue(domain.Name))
			})
			if len(domains) == 0 {
				return fmt.Errorf("could not find any domains matched by certificate")
			}
		}

	default:
		return fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 批量更新域名证书
	if len(domains) == 0 {
		d.logger.Info("no mse cloud-native domains to deploy")
	} else {
		domainnames := lo.Map(domains, func(domain *alimse.ListGatewayDomainResponseBodyData, _ int) string {
			return tea.StringValue(domain.Name)
		})
		d.logger.Info("found mse cloud-native domains to deploy", slog.Any("domains", domainnames))

		if err := xloop.ForRangeAllWithContext(ctx, domains, func(ctx context.Context, domain *alimse.ListGatewayDomainResponseBodyData, _ int) error {
			certId := upres.ExtendedData["CertIdWithRegion"].(string)
			return d.updateCloudNativeDomainCertificate(ctx, d.config.GatewayId, domain, certId)
		}); err != nil {
			return err
		}
	}

	return nil
}

func (d *Deployer) getCloudNativeAllDomainsByGatewayId(ctx context.Context, cloudGatewayId string) ([]*alimse.ListGatewayDomainResponseBodyData, error) {
	domains := make([]*alimse.ListGatewayDomainResponseBodyData, 0)

	// 查询网关已关联域名列表
	// REF: https://help.aliyun.com/zh/mse/developer-reference/api-mse-2019-05-31-listgatewaydomain
	listGatewayDomainReq := &alimse.ListGatewayDomainRequest{
		GatewayUniqueId: tea.String(cloudGatewayId),
	}
	listGatewayDomainResp, err := d.sdkClient.ListGatewayDomainWithContext(ctx, listGatewayDomainReq, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'mse.ListGatewayDomain'", slog.Any("request", listGatewayDomainReq), slog.Any("response", listGatewayDomainResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'mse.ListGatewayDomain': %w", err)
	} else if !tea.BoolValue(listGatewayDomainResp.Body.Success) {
		return nil, fmt.Errorf("failed to execute sdk request 'mse.ListGatewayDomain': [%d] %s", tea.Int32Value(listGatewayDomainResp.Body.Code), tea.StringValue(listGatewayDomainResp.Body.Message))
	}

	ignoredStatuses := []int32{5, 6}
	for _, domainItem := range listGatewayDomainResp.Body.Data {
		if lo.Contains(ignoredStatuses, tea.Int32Value(domainItem.Status)) {
			continue
		}

		domains = append(domains, domainItem)
	}

	return domains, nil
}

func (d *Deployer) updateCloudNativeDomainCertificate(ctx context.Context, cloudGatewayId string, cloudDomainInfo *alimse.ListGatewayDomainResponseBodyData, cloudCertId string) error {
	if strings.EqualFold(tea.StringValue(cloudDomainInfo.Protocol), "HTTPS") {
		// 已部署过，直接返回
		if tea.StringValue(cloudDomainInfo.CertIdentifier) == cloudCertId {
			return nil
		}

		// 更换证书
		// REF: https://help.aliyun.com/zh/mse/developer-reference/api-mse-2019-05-31-updatesslcert
		updateSSLCertReq := &alimse.UpdateSSLCertRequest{
			GatewayUniqueId: tea.String(cloudGatewayId),
			DomainId:        cloudDomainInfo.Id,
			CertIdentifier:  tea.String(cloudCertId),
		}
		updateSSLCertResp, err := d.sdkClient.UpdateSSLCertWithContext(ctx, updateSSLCertReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'mse.UpdateSSLCert'", slog.Any("request", updateSSLCertReq), slog.Any("response", updateSSLCertResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'mse.UpdateSSLCert': %w", err)
		} else if !tea.BoolValue(updateSSLCertResp.Body.Success) {
			return fmt.Errorf("failed to execute sdk request 'mse.UpdateSSLCert': [%d] %s", tea.Int32Value(updateSSLCertResp.Body.Code), tea.StringValue(updateSSLCertResp.Body.Message))
		}

		return nil
	}

	// 更新关联域名信息
	// REF: https://help.aliyun.com/zh/mse/developer-reference/api-mse-2019-05-31-updategatewaydomain
	updateGatewayDomainReq := &alimse.UpdateGatewayDomainRequest{
		GatewayUniqueId: tea.String(cloudGatewayId),
		Id:              cloudDomainInfo.Id,
		Protocol:        tea.String("HTTPS"),
		CertIdentifier:  tea.String(cloudCertId),
		MustHttps:       tea.Bool(false),
		TlsMin:          tea.String(lo.CoalesceOrEmpty(tea.StringValue(cloudDomainInfo.TlsMin), "TLS 1.2")),
		TlsMax:          tea.String(lo.CoalesceOrEmpty(tea.StringValue(cloudDomainInfo.TlsMax), "TLS 1.3")),
	}
	updateGatewayDomainResp, err := d.sdkClient.UpdateGatewayDomainWithContext(ctx, updateGatewayDomainReq, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'mse.UpdateGatewayDomain'", slog.Any("request", updateGatewayDomainReq), slog.Any("response", updateGatewayDomainResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'mse.UpdateGatewayDomain': %w", err)
	} else if !tea.BoolValue(updateGatewayDomainResp.Body.Success) {
		return fmt.Errorf("failed to execute sdk request 'mse.UpdateGatewayDomain': [%d] %s", tea.Int32Value(updateGatewayDomainResp.Body.Code), tea.StringValue(updateGatewayDomainResp.Body.Message))
	}

	return nil
}

func createSDKClient(accessKeyId, accessKeySecret, region string) (*alimse.Client, error) {
	// 接入点一览 https://api.aliyun.com/product/Mse
	endpoint := fmt.Sprintf("mse.%s.aliyuncs.com", region)

	config := &aliopen.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(endpoint),
	}

	client, err := alimse.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
