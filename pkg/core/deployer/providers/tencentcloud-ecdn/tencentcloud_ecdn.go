package tencentcloudecdn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/samber/lo"
	tccdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/tencentcloud-ssl"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-ecdn/internal"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
)

type DeployerConfig struct {
	// 腾讯云 SecretId。
	SecretId string `json:"secretId"`
	// 腾讯云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 腾讯云接口端点。
	Endpoint string `json:"endpoint,omitempty"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 加速域名（支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *internal.CdnClient
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.SecretId, config.SecretKey, config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		SecretId:  config.SecretId,
		SecretKey: config.SecretKey,
		Endpoint: lo.
			If(strings.HasSuffix(config.Endpoint, "intl.tencentcloudapi.com"), "ssl.intl.tencentcloudapi.com"). // 国际站使用独立的接口端点
			Else(""),
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

	// 获取待部署的 ECDN 实例
	domains := make([]string, 0)
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if d.config.Domain == "" {
				return nil, errors.New("config `domain` is required")
			}

			domains = []string{d.config.Domain}
		}

	case DOMAIN_MATCH_PATTERN_WILDCARD:
		{
			if d.config.Domain == "" {
				return nil, errors.New("config `domain` is required")
			}

			if strings.HasPrefix(d.config.Domain, "*.") {
				domainCandidates, err := d.getMatchedDomainsByWildcard(ctx, d.config.Domain)
				if err != nil {
					return nil, err
				}

				domains = domainCandidates
			} else {
				domains = []string{d.config.Domain}
			}
		}

	case DOMAIN_MATCH_PATTERN_CERTSAN:
		{
			domainCandidates, err := d.getMatchedDomainsByCertId(ctx, upres.CertId)
			if err != nil {
				return nil, err
			}

			domains = domainCandidates
		}

	default:
		return nil, fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 遍历更新域名证书
	if len(domains) == 0 {
		d.logger.Info("no ecdn domains to deploy")
	} else {
		d.logger.Info("found ecdn domains to deploy", slog.Any("domains", domains))
		var errs []error

		for _, domain := range domains {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				if err := d.updateDomainCertificate(ctx, domain, upres.CertId); err != nil {
					errs = append(errs, err)
				}
			}
		}

		if len(errs) > 0 {
			return nil, errors.Join(errs...)
		}
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) getMatchedDomainsByWildcard(ctx context.Context, wildcardDomain string) ([]string, error) {
	domains := make([]string, 0)

	// 查询域名基本信息，获取匹配的域名
	// REF: https://cloud.tencent.com/document/api/228/41118
	describeDomainsOffset := 0
	describeDomainsLimit := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		describeDomainsReq := tccdn.NewDescribeDomainsRequest()
		describeDomainsReq.Filters = []*tccdn.DomainFilter{
			{
				Name:  common.StringPtr("domain"),
				Value: common.StringPtrs([]string{strings.TrimPrefix(wildcardDomain, "*.")}),
				Fuzzy: common.BoolPtr(true),
			},
		}
		describeDomainsReq.Offset = common.Int64Ptr(int64(describeDomainsOffset))
		describeDomainsReq.Limit = common.Int64Ptr(int64(describeDomainsLimit))
		describeDomainsResp, err := d.sdkClient.DescribeDomains(describeDomainsReq)
		d.logger.Debug("sdk request 'cdn.DescribeDomains'", slog.Any("request", describeDomainsReq), slog.Any("response", describeDomainsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'cdn.DescribeDomains': %w", err)
		}

		if describeDomainsResp.Response == nil {
			break
		}

		for _, domainItem := range describeDomainsResp.Response.Domains {
			if lo.FromPtr(domainItem.Product) == "ecdn" && xcerthostname.IsMatch(wildcardDomain, lo.FromPtr(domainItem.Domain)) {
				domains = append(domains, *domainItem.Domain)
			}
		}

		if len(describeDomainsResp.Response.Domains) < describeDomainsLimit {
			break
		}

		describeDomainsOffset += describeDomainsLimit
	}

	return domains, nil
}

func (d *Deployer) getMatchedDomainsByCertId(ctx context.Context, cloudCertId string) ([]string, error) {
	// 获取证书中的可用域名
	// REF: https://cloud.tencent.com/document/api/228/42491
	describeCertDomainsReq := tccdn.NewDescribeCertDomainsRequest()
	describeCertDomainsReq.CertId = common.StringPtr(cloudCertId)
	describeCertDomainsReq.Product = common.StringPtr("ecdn")
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

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain string, cloudCertId string) error {
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
		return fmt.Errorf("could not find domain '%s'", domain)
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

func createSDKClient(secretId, secretKey, endpoint string) (*internal.CdnClient, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := internal.NewCdnClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
