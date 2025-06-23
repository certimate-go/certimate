package tencentcloudcdn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	tccdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"

	"github.com/certimate-go/certimate/pkg/core"
	sslmgrsp "github.com/certimate-go/certimate/pkg/core/ssl-manager/providers/tencentcloud-ssl"
	"github.com/certimate-go/certimate/pkg/utils/ifelse"
)

type SSLDeployerProviderConfig struct {
	// 腾讯云 SecretId。
	SecretId string `json:"secretId"`
	// 腾讯云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 腾讯云接口端点。
	Endpoint string `json:"endpoint,omitempty"`
	// 加速域名（支持泛域名）。
	Domain string `json:"domain"`
}

type SSLDeployerProvider struct {
	config     *SSLDeployerProviderConfig
	logger     *slog.Logger
	sdkClient  *tccdn.Client
	sslManager core.SSLManager
}

var _ core.SSLDeployer = (*SSLDeployerProvider)(nil)

type wSDKClients struct {
	SSL *tcssl.Client
	CDN *tccdn.Client
}

func NewSSLDeployerProvider(config *SSLDeployerProviderConfig) (*SSLDeployerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl deployer provider is nil")
	}

	client, err := createSDKClient(config.SecretId, config.SecretKey, config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("could not create sdk client: %w", err)
	}

	sslmgr, err := sslmgrsp.NewSSLManagerProvider(&sslmgrsp.SSLManagerProviderConfig{
		SecretId:  config.SecretId,
		SecretKey: config.SecretKey,
		Endpoint: ifelse.
			If[string](strings.HasSuffix(strings.TrimSpace(config.Endpoint), "intl.tencentcloudapi.com")).
			Then("ssl.intl.tencentcloudapi.com"). // 国际站使用独立的接口端点
			Else(""),
	})
	if err != nil {
		return nil, fmt.Errorf("could not create ssl manager: %w", err)
	}

	return &SSLDeployerProvider{
		config:     config,
		logger:     slog.Default(),
		sdkClient:  client,
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
	if d.config.Domain == "" {
		return nil, errors.New("config `domain` is required")
	}

	// 上传证书
	upres, err := d.sslManager.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 获取待部署的 CDN 实例
	// 如果是泛域名，根据证书匹配 CDN 实例
	domains := make([]string, 0)
	if strings.HasPrefix(d.config.Domain, "*.") {
		temp, err := d.getDomainsByCertId(ctx, upres.CertId)
		if err != nil {
			return nil, err
		}

		domains = temp
	} else {
		domains = append(domains, d.config.Domain)
	}

	// 遍历更新域名证书
	if len(domains) == 0 {
		d.logger.Info("no cdn domains to deploy")
	} else {
		d.logger.Info("found cdn domains to deploy", slog.Any("domains", domains))
		var errs []error

		for _, domain := range domains {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				if err := d.updateDomainHttpsServerCert(ctx, domain, upres.CertId); err != nil {
					errs = append(errs, err)
				}
			}
		}

		if len(errs) > 0 {
			return nil, errors.Join(errs...)
		}
	}

	return &core.SSLDeployResult{}, nil
}

func (d *SSLDeployerProvider) getDomainsByCertId(ctx context.Context, cloudCertId string) ([]string, error) {
	// 获取证书中的可用域名
	// REF: https://cloud.tencent.com/document/api/228/42491
	describeCertDomainsReq := tccdn.NewDescribeCertDomainsRequest()
	describeCertDomainsReq.CertId = common.StringPtr(cloudCertId)
	describeCertDomainsReq.Product = common.StringPtr("cdn")
	describeCertDomainsResp, err := d.sdkClient.DescribeCertDomains(describeCertDomainsReq)
	d.logger.Debug("sdk request 'cdn.DescribeCertDomains'", slog.Any("request", describeCertDomainsReq), slog.Any("response", describeCertDomainsResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cdn.DescribeCertDomains': %w", err)
	}

	domains := make([]string, 0)
	if describeCertDomainsResp.Response.Domains != nil {
		for _, domain := range describeCertDomainsResp.Response.Domains {
			domains = append(domains, *domain)
		}
	}

	return domains, nil
}

func (d *SSLDeployerProvider) updateDomainHttpsServerCert(ctx context.Context, domain string, cloudCertId string) error {
	// 查询域名详细配置
	// REF: https://cloud.tencent.com/document/api/228/41117
	describeDomainsConfigReq := tccdn.NewDescribeDomainsConfigRequest()
	describeDomainsConfigReq.Filters = []*tccdn.DomainFilter{
		{
			Name:  common.StringPtr("domain"),
			Value: common.StringPtrs([]string{domain}),
		},
	}
	describeDomainsConfigReq.Offset = common.Int64Ptr(0)
	describeDomainsConfigReq.Limit = common.Int64Ptr(1)
	describeDomainsConfigResp, err := d.sdkClient.DescribeDomainsConfig(describeDomainsConfigReq)
	d.logger.Debug("sdk request 'cdn.DescribeDomainsConfig'", slog.Any("request", describeDomainsConfigReq), slog.Any("response", describeDomainsConfigResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'cdn.DescribeDomainsConfig': %w", err)
	} else if len(describeDomainsConfigResp.Response.Domains) == 0 {
		return fmt.Errorf("domain %s not found", domain)
	}

	domainConfig := describeDomainsConfigResp.Response.Domains[0]
	if domainConfig.Https != nil && domainConfig.Https.CertInfo != nil && domainConfig.Https.CertInfo.CertId != nil && *domainConfig.Https.CertInfo.CertId == cloudCertId {
		// 已部署过此域名，跳过
		return nil
	}

	// 更新加速域名配置
	// REF: https://cloud.tencent.com/document/api/228/41116
	updateDomainConfigReq := tccdn.NewUpdateDomainConfigRequest()
	updateDomainConfigReq.Domain = common.StringPtr(domain)
	updateDomainConfigReq.Https = domainConfig.Https
	if updateDomainConfigReq.Https == nil {
		updateDomainConfigReq.Https = &tccdn.Https{
			Switch: common.StringPtr("on"),
		}
	} else {
		updateDomainConfigReq.Https.SslStatus = nil
	}
	updateDomainConfigReq.Https.CertInfo = &tccdn.ServerCert{
		CertId: common.StringPtr(cloudCertId),
	}
	updateDomainConfigResp, err := d.sdkClient.UpdateDomainConfig(updateDomainConfigReq)
	d.logger.Debug("sdk request 'cdn.UpdateDomainConfig'", slog.Any("request", updateDomainConfigReq), slog.Any("response", updateDomainConfigResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'cdn.UpdateDomainConfig': %w", err)
	}

	return nil
}

func createSDKClient(secretId, secretKey, endpoint string) (*tccdn.Client, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := tccdn.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
