package aliyunapigw

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	aliapig "github.com/alibabacloud-go/apig-20240327/v6/client"
	alicloudapi "github.com/alibabacloud-go/cloudapi-20160714/v5/client"
	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aliyun-cas"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-apigw/internal"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
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
	// API 网关 ID。
	// 服务类型为 [SERVICE_TYPE_CLOUDNATIVE] 时必填。
	GatewayId string `json:"gatewayId,omitempty"`
	// API 分组 ID。
	// 服务类型为 [SERVICE_TYPE_TRADITIONAL] 时必填。
	GroupId string `json:"groupId,omitempty"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 自定义域名（支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClients *wSDKClients
	sdkCertmgr certmgr.Provider
}

type wSDKClients struct {
	CloudNativeAPIGateway *internal.ApigClient
	TraditionalAPIGateway *internal.CloudapiClient
}

var _ deployer.Provider = (*Deployer)(nil)

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
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM, issuerCertificatePEM string) (*deployer.DeployResult, error) {
	switch d.config.ServiceType {
	case SERVICE_TYPE_TRADITIONAL:
		if err := d.deployToTraditional(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	case SERVICE_TYPE_CLOUDNATIVE:
		if err := d.deployToCloudNative(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported service type '%s'", string(d.config.ServiceType))
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) deployToTraditional(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.GroupId == "" {
		return errors.New("config `groupId` is required")
	}

	// 获取待部署的域名列表
	var domains []string
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if d.config.Domain == "" {
				return errors.New("config `domain` is required")
			}

			domains = []string{d.config.Domain}
		}

	case DOMAIN_MATCH_PATTERN_WILDCARD:
		{
			if d.config.Domain == "" {
				return errors.New("config `domain` is required")
			}

			if strings.HasPrefix(d.config.Domain, "*.") {
				domainCandidates, err := d.getTraditionalAllDomainsByGroupId(ctx, d.config.GroupId)
				if err != nil {
					return err
				}

				domains = lo.Filter(domainCandidates, func(domain string, _ int) bool {
					return xcerthostname.IsMatch(d.config.Domain, domain)
				})
				if len(domains) == 0 {
					return errors.New("could not find any domains matched by wildcard")
				}
			} else {
				domains = []string{d.config.Domain}
			}
		}

	case DOMAIN_MATCH_PATTERN_CERTSAN:
		{
			domainCandidates, err := d.getTraditionalAllDomainsByGroupId(ctx, d.config.GroupId)
			if err != nil {
				return err
			}

			domains = lo.Filter(domainCandidates, func(domain string, _ int) bool {
				return xcerthostname.IsMatchByCertificatePEM(certPEM, domain)
			})
			if len(domains) == 0 {
				return errors.New("could not find any domains matched by certificate")
			}
		}

	default:
		return fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 遍历更新域名证书
	if len(domains) == 0 {
		d.logger.Info("no apigw domains to deploy")
	} else {
		d.logger.Info("found apigw domains to deploy", slog.Any("domains", domains))
		var errs []error

		for _, domain := range domains {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := d.updateTraditionalDomainCertificate(ctx, d.config.GroupId, domain, certPEM, privkeyPEM); err != nil {
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

func (d *Deployer) deployToCloudNative(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.GatewayId == "" {
		return errors.New("config `gatewayId` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 获取待部署的域名列表
	var domains []string
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if d.config.Domain == "" {
				return errors.New("config `domain` is required")
			}

			domains = []string{d.config.Domain}
		}

	case DOMAIN_MATCH_PATTERN_WILDCARD:
		{
			if d.config.Domain == "" {
				return errors.New("config `domain` is required")
			}

			if strings.HasPrefix(d.config.Domain, "*.") {
				domainCandidates, err := d.getCloudNativeAllDomainsByGatewayId(ctx, d.config.GatewayId)
				if err != nil {
					return err
				}

				domains = lo.Filter(domainCandidates, func(domain string, _ int) bool {
					return xcerthostname.IsMatch(d.config.Domain, domain)
				})
				if len(domains) == 0 {
					return errors.New("could not find any domains matched by wildcard")
				}
			} else {
				domains = []string{d.config.Domain}
			}
		}

	case DOMAIN_MATCH_PATTERN_CERTSAN:
		{
			domainCandidates, err := d.getCloudNativeAllDomainsByGatewayId(ctx, d.config.GatewayId)
			if err != nil {
				return err
			}

			domains = lo.Filter(domainCandidates, func(domain string, _ int) bool {
				return xcerthostname.IsMatchByCertificatePEM(certPEM, domain)
			})
			if len(domains) == 0 {
				return errors.New("could not find any domains matched by certificate")
			}
		}

	default:
		return fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 遍历更新域名证书
	if len(domains) == 0 {
		d.logger.Info("no apigw domains to deploy")
	} else {
		d.logger.Info("found apigw domains to deploy", slog.Any("domains", domains))
		var errs []error

		for _, domain := range domains {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				certId := upres.ExtendedData["CertIdentifier"].(string)
				if err := d.updateCloudNativeDomainCertificate(ctx, d.config.GatewayId, domain, certId); err != nil {
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

func (d *Deployer) getTraditionalAllDomainsByGroupId(ctx context.Context, cloudGroupId string) ([]string, error) {
	domains := make([]string, 0)

	// 查询 API 分组详情
	// REF: https://help.aliyun.com/zh/api-gateway/traditional-api-gateway/developer-reference/api-cloudapi-2016-07-14-describeapigroup
	describeApiGroupReq := &alicloudapi.DescribeApiGroupRequest{
		GroupId: tea.String(cloudGroupId),
	}
	describeApiGroupResp, err := d.sdkClients.TraditionalAPIGateway.DescribeApiGroupWithContext(ctx, describeApiGroupReq, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'apigateway.DescribeApiGroup'", slog.Any("request", describeApiGroupReq), slog.Any("response", describeApiGroupResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'apigateway.DescribeApiGroup': %w", err)
	}

	for _, domainItem := range describeApiGroupResp.Body.CustomDomains.DomainItem {
		if strings.EqualFold(tea.StringValue(domainItem.DomainBindingStatus), "BINDING") {
			domains = append(domains, tea.StringValue(domainItem.DomainName))
		}
	}

	return domains, nil
}

func (d *Deployer) getCloudNativeAllDomainsByGatewayId(ctx context.Context, cloudGatewayId string) ([]string, error) {
	domains := make([]string, 0)

	// 查询域名列表
	// REF: https://help.aliyun.com/zh/api-gateway/cloud-native-api-gateway/developer-reference/api-apig-2024-03-27-listdomains
	listDomainsPageNumber := 1
	listDomainsPageSize := 10
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listDomainsReq := &aliapig.ListDomainsRequest{
			ResourceGroupId: lo.EmptyableToPtr(d.config.ResourceGroupId),
			GatewayId:       tea.String(cloudGatewayId),
			PageNumber:      tea.Int32(int32(listDomainsPageNumber)),
			PageSize:        tea.Int32(int32(listDomainsPageSize)),
		}
		listDomainsResp, err := d.sdkClients.CloudNativeAPIGateway.ListDomainsWithContext(ctx, listDomainsReq, make(map[string]*string), &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'apig.ListDomains'", slog.Any("request", listDomainsReq), slog.Any("response", listDomainsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'apig.ListDomains': %w", err)
		}

		if listDomainsResp.Body == nil || listDomainsResp.Body.Data == nil {
			break
		}

		for _, domainItem := range listDomainsResp.Body.Data.Items {
			if strings.EqualFold(tea.StringValue(domainItem.Status), "Published") {
				domains = append(domains, tea.StringValue(domainItem.Name))
			}
		}

		if len(listDomainsResp.Body.Data.Items) < listDomainsPageSize {
			break
		}

		listDomainsPageNumber++
	}

	return domains, nil
}

func (d *Deployer) updateTraditionalDomainCertificate(ctx context.Context, cloudGroupId string, domain string, certPEM, privkeyPEM string) error {
	// 为自定义域名添加 SSL 证书
	// REF: https://help.aliyun.com/zh/api-gateway/traditional-api-gateway/developer-reference/api-cloudapi-2016-07-14-setdomaincertificate
	setDomainCertificateReq := &alicloudapi.SetDomainCertificateRequest{
		GroupId:               tea.String(cloudGroupId),
		DomainName:            tea.String(domain),
		CertificateName:       tea.String(fmt.Sprintf("certimate_%d", time.Now().UnixMilli())),
		CertificateBody:       tea.String(certPEM),
		CertificatePrivateKey: tea.String(privkeyPEM),
	}
	setDomainCertificateResp, err := d.sdkClients.TraditionalAPIGateway.SetDomainCertificateWithContext(ctx, setDomainCertificateReq, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'apigateway.SetDomainCertificate'", slog.Any("request", setDomainCertificateReq), slog.Any("response", setDomainCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'apigateway.SetDomainCertificate': %w", err)
	}

	return nil
}

func (d *Deployer) updateCloudNativeDomainCertificate(ctx context.Context, cloudGatewayId string, domain string, cloudCertId string) error {
	// 获取域名 ID
	domainId, err := d.findCloudNativeDomainIdByDomain(ctx, cloudGatewayId, domain)
	if err != nil {
		return err
	}

	// 查询域名
	// REF: https://help.aliyun.com/zh/api-gateway/cloud-native-api-gateway/developer-reference/api-apig-2024-03-27-getdomain
	getDomainReq := &aliapig.GetDomainRequest{}
	getDomainResp, err := d.sdkClients.CloudNativeAPIGateway.GetDomainWithContext(ctx, tea.String(domainId), getDomainReq, make(map[string]*string), &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'apig.GetDomain'", slog.String("domainId", domainId), slog.Any("request", getDomainReq), slog.Any("response", getDomainResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'apig.GetDomain': %w", err)
	}

	// 更新域名
	// REF: https://help.aliyun.com/zh/api-gateway/cloud-native-api-gateway/developer-reference/api-apig-2024-03-27-updatedomain
	updateDomainReq := &aliapig.UpdateDomainRequest{
		Protocol:              tea.String("HTTPS"),
		ForceHttps:            getDomainResp.Body.Data.ForceHttps,
		MTLSEnabled:           getDomainResp.Body.Data.MTLSEnabled,
		Http2Option:           getDomainResp.Body.Data.Http2Option,
		TlsMin:                getDomainResp.Body.Data.TlsMin,
		TlsMax:                getDomainResp.Body.Data.TlsMax,
		TlsCipherSuitesConfig: getDomainResp.Body.Data.TlsCipherSuitesConfig,
		CertIdentifier:        tea.String(cloudCertId),
	}
	updateDomainResp, err := d.sdkClients.CloudNativeAPIGateway.UpdateDomainWithContext(ctx, tea.String(domainId), updateDomainReq, make(map[string]*string), &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'apig.UpdateDomain'", slog.String("domainId", domainId), slog.Any("request", updateDomainReq), slog.Any("response", updateDomainResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'apig.UpdateDomain': %w", err)
	}

	return nil
}

func (d *Deployer) findCloudNativeDomainIdByDomain(ctx context.Context, cloudGatewayId string, domain string) (string, error) {
	// 查询域名列表
	// REF: https://help.aliyun.com/zh/api-gateway/cloud-native-api-gateway/developer-reference/api-apig-2024-03-27-listdomains
	listDomainsPageNumber := 1
	listDomainsPageSize := 10
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		listDomainsReq := &aliapig.ListDomainsRequest{
			ResourceGroupId: lo.EmptyableToPtr(d.config.ResourceGroupId),
			GatewayId:       tea.String(cloudGatewayId),
			NameLike:        tea.String(domain),
			PageNumber:      tea.Int32(int32(listDomainsPageNumber)),
			PageSize:        tea.Int32(int32(listDomainsPageSize)),
		}
		listDomainsResp, err := d.sdkClients.CloudNativeAPIGateway.ListDomainsWithContext(ctx, listDomainsReq, make(map[string]*string), &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'apig.ListDomains'", slog.Any("request", listDomainsReq), slog.Any("response", listDomainsResp))
		if err != nil {
			return "", fmt.Errorf("failed to execute sdk request 'apig.ListDomains': %w", err)
		}

		if listDomainsResp.Body == nil || listDomainsResp.Body.Data == nil {
			break
		}

		for _, domainItem := range listDomainsResp.Body.Data.Items {
			if strings.EqualFold(tea.StringValue(domainItem.Name), domain) {
				return tea.StringValue(domainItem.DomainId), nil
			}
		}

		if len(listDomainsResp.Body.Data.Items) < listDomainsPageSize {
			break
		}

		listDomainsPageNumber++
	}

	return "", fmt.Errorf("could not find domain '%s'", domain)
}

func createSDKClients(accessKeyId, accessKeySecret, region string) (*wSDKClients, error) {
	// 接入点一览 https://api.aliyun.com/product/APIG
	var cloudNativeAPIGEndpoint string
	switch region {
	case "":
		cloudNativeAPIGEndpoint = "apig.cn-hangzhou.aliyuncs.com"
	default:
		cloudNativeAPIGEndpoint = fmt.Sprintf("apig.%s.aliyuncs.com", region)
	}

	cloudNativeAPIGConfig := &aliopen.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(cloudNativeAPIGEndpoint),
	}
	cloudNativeAPIGClient, err := internal.NewApigClient(cloudNativeAPIGConfig)
	if err != nil {
		return nil, err
	}

	// 接入点一览 https://api.aliyun.com/product/CloudAPI
	var traditionalAPIGEndpoint string
	switch region {
	case "":
		traditionalAPIGEndpoint = "apigateway.cn-hangzhou.aliyuncs.com"
	default:
		traditionalAPIGEndpoint = fmt.Sprintf("apigateway.%s.aliyuncs.com", region)
	}

	traditionalAPIGConfig := &aliopen.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(traditionalAPIGEndpoint),
	}
	traditionalAPIGClient, err := internal.NewCloudapiClient(traditionalAPIGConfig)
	if err != nil {
		return nil, err
	}

	return &wSDKClients{
		CloudNativeAPIGateway: cloudNativeAPIGClient,
		TraditionalAPIGateway: traditionalAPIGClient,
	}, nil
}
