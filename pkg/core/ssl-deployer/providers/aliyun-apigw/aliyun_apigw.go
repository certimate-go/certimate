package aliyunapigw

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	aliapig "github.com/alibabacloud-go/apig-20240327/v3/client"
	alicloudapi "github.com/alibabacloud-go/cloudapi-20160714/v5/client"
	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/certimate-go/certimate/pkg/core"
	sslmgrsp "github.com/certimate-go/certimate/pkg/core/ssl-manager/providers/aliyun-cas"
	"github.com/certimate-go/certimate/pkg/utils/ifelse"
	xtypes "github.com/certimate-go/certimate/pkg/utils/types"
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
	// 服务类型。
	ServiceType ServiceType `json:"serviceType"`
	// API 网关 ID。
	// 服务类型为 [SERVICE_TYPE_CLOUDNATIVE] 时必填。
	GatewayId string `json:"gatewayId,omitempty"`
	// API 分组 ID。
	// 服务类型为 [SERVICE_TYPE_TRADITIONAL] 时必填。
	GroupId string `json:"groupId,omitempty"`
	// 自定义域名（支持泛域名）。
	Domain string `json:"domain"`
}

type SSLDeployerProvider struct {
	config     *SSLDeployerProviderConfig
	logger     *slog.Logger
	sdkClients *wSDKClients
	sslManager core.SSLManager
}

type wSDKClients struct {
	CloudNativeAPIGateway *aliapig.Client
	TraditionalAPIGateway *alicloudapi.Client
}

var _ core.SSLDeployer = (*SSLDeployerProvider)(nil)

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
}

func (d *SSLDeployerProvider) Deploy(ctx context.Context, certPEM string, privkeyPEM string) (*core.SSLDeployResult, error) {
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

	return &core.SSLDeployResult{}, nil
}

func (d *SSLDeployerProvider) deployToTraditional(ctx context.Context, certPEM string, privkeyPEM string) error {
	if d.config.GroupId == "" {
		return errors.New("config `groupId` is required")
	}
	if d.config.Domain == "" {
		return errors.New("config `domain` is required")
	}

	// 为自定义域名添加 SSL 证书
	// REF: https://help.aliyun.com/zh/api-gateway/traditional-api-gateway/developer-reference/api-cloudapi-2016-07-14-setdomaincertificate
	setDomainCertificateReq := &alicloudapi.SetDomainCertificateRequest{
		GroupId:               tea.String(d.config.GroupId),
		DomainName:            tea.String(d.config.Domain),
		CertificateName:       tea.String(fmt.Sprintf("certimate_%d", time.Now().UnixMilli())),
		CertificateBody:       tea.String(certPEM),
		CertificatePrivateKey: tea.String(privkeyPEM),
	}
	setDomainCertificateResp, err := d.sdkClients.TraditionalAPIGateway.SetDomainCertificate(setDomainCertificateReq)
	d.logger.Debug("sdk request 'apigateway.SetDomainCertificate'", slog.Any("request", setDomainCertificateReq), slog.Any("response", setDomainCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'apigateway.SetDomainCertificate': %w", err)
	}

	return nil
}

func (d *SSLDeployerProvider) deployToCloudNative(ctx context.Context, certPEM string, privkeyPEM string) error {
	if d.config.GatewayId == "" {
		return errors.New("config `gatewayId` is required")
	}
	if d.config.Domain == "" {
		return errors.New("config `domain` is required")
	}

	// 遍历查询域名列表，获取域名 ID
	// REF: https://help.aliyun.com/zh/api-gateway/cloud-native-api-gateway/developer-reference/api-apig-2024-03-27-listdomains
	var domainId string
	listDomainsPageNumber := int32(1)
	listDomainsPageSize := int32(10)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		listDomainsReq := &aliapig.ListDomainsRequest{
			ResourceGroupId: xtypes.ToPtrOrZeroNil(d.config.ResourceGroupId),
			GatewayId:       tea.String(d.config.GatewayId),
			NameLike:        tea.String(d.config.Domain),
			PageNumber:      tea.Int32(listDomainsPageNumber),
			PageSize:        tea.Int32(listDomainsPageSize),
		}
		listDomainsResp, err := d.sdkClients.CloudNativeAPIGateway.ListDomains(listDomainsReq)
		d.logger.Debug("sdk request 'apig.ListDomains'", slog.Any("request", listDomainsReq), slog.Any("response", listDomainsResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'apig.ListDomains': %w", err)
		}

		if listDomainsResp.Body.Data.Items != nil {
			for _, domainInfo := range listDomainsResp.Body.Data.Items {
				if strings.EqualFold(tea.StringValue(domainInfo.Name), d.config.Domain) {
					domainId = tea.StringValue(domainInfo.DomainId)
					break
				}
			}

			if domainId != "" {
				break
			}
		}

		if listDomainsResp.Body.Data.Items == nil || len(listDomainsResp.Body.Data.Items) < int(listDomainsPageSize) {
			break
		} else {
			listDomainsPageNumber++
		}
	}
	if domainId == "" {
		return errors.New("domain not found")
	}

	// 查询域名
	// REF: https://help.aliyun.com/zh/api-gateway/cloud-native-api-gateway/developer-reference/api-apig-2024-03-27-getdomain
	getDomainReq := &aliapig.GetDomainRequest{}
	getDomainResp, err := d.sdkClients.CloudNativeAPIGateway.GetDomain(tea.String(domainId), getDomainReq)
	d.logger.Debug("sdk request 'apig.GetDomain'", slog.Any("domainId", domainId), slog.Any("request", getDomainReq), slog.Any("response", getDomainResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'apig.GetDomain': %w", err)
	}

	// 上传证书
	upres, err := d.sslManager.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
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
		CertIdentifier:        tea.String(upres.ExtendedData["certIdentifier"].(string)),
	}
	updateDomainResp, err := d.sdkClients.CloudNativeAPIGateway.UpdateDomain(tea.String(domainId), updateDomainReq)
	d.logger.Debug("sdk request 'apig.UpdateDomain'", slog.Any("domainId", domainId), slog.Any("request", updateDomainReq), slog.Any("response", updateDomainResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'apig.UpdateDomain': %w", err)
	}

	return nil
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
	cloudNativeAPIGClient, err := aliapig.NewClient(cloudNativeAPIGConfig)
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
	traditionalAPIGClient, err := alicloudapi.NewClient(traditionalAPIGConfig)
	if err != nil {
		return nil, err
	}

	return &wSDKClients{
		CloudNativeAPIGateway: cloudNativeAPIGClient,
		TraditionalAPIGateway: traditionalAPIGClient,
	}, nil
}
