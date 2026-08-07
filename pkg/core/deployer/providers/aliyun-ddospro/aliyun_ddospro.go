package aliyunddospro

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/samber/lo"

	aliddoscoo "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/alibabacloud-go/ddoscoo-20200101/v5/client"

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
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 网站域名（支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *aliddoscoo.Client
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
	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 获取待部署的域名列表
	var domains []string
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if d.config.Domain == "" {
				return nil, fmt.Errorf("config `domain` is required")
			}

			domains = []string{d.config.Domain}
		}

	case DOMAIN_MATCH_PATTERN_WILDCARD:
		{
			if d.config.Domain == "" {
				return nil, fmt.Errorf("config `domain` is required")
			}

			if strings.HasPrefix(d.config.Domain, "*.") {
				domainCandidates, err := d.getAllDomains(ctx)
				if err != nil {
					return nil, err
				}

				domains = lo.Filter(domainCandidates, func(domain string, _ int) bool {
					return xcerthostname.IsMatch(d.config.Domain, domain)
				})
				if len(domains) == 0 {
					return nil, fmt.Errorf("could not find any domains matched by wildcard")
				}
			} else {
				domains = []string{d.config.Domain}
			}
		}

	case DOMAIN_MATCH_PATTERN_CERTSAN:
		{
			domainCandidates, err := d.getAllDomains(ctx)
			if err != nil {
				return nil, err
			}

			domains = lo.Filter(domainCandidates, func(domain string, _ int) bool {
				return xcerthostname.IsMatchByCertificatePEM(certPEM, domain)
			})
			if len(domains) == 0 {
				return nil, fmt.Errorf("could not find any domains matched by certificate")
			}
		}

	default:
		return nil, fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 批量更新域名证书
	if len(domains) == 0 {
		d.logger.Info("no ddoscoo domains to deploy")
	} else {
		d.logger.Info("found ddoscoo domains to deploy", slog.Any("domains", domains))

		if err := xloop.ForRangeAllWithContext(ctx, domains, func(ctx context.Context, domain string, _ int) error {
			certId := upres.ExtendedData["CertIdWithRegion"].(string)
			return d.updateDomainCertificate(ctx, domain, certId)
		}); err != nil {
			return nil, err
		}
	}

	return &DeployResult{}, nil
}

func (d *Deployer) getAllDomains(ctx context.Context) ([]string, error) {
	domains := make([]string, 0)

	// 查询已配置网站业务转发规则的域名
	// REF: https://help.aliyun.com/zh/anti-ddos/anti-ddos-pro-and-premium/developer-reference/api-ddoscoo-2020-01-01-describedomains
	describeDomainsReq := &aliddoscoo.DescribeDomainsRequest{
		ResourceGroupId: lo.EmptyableToPtr(d.config.ResourceGroupId),
	}
	describeDomainsResp, err := d.sdkClient.DescribeDomainsWithContext(ctx, describeDomainsReq, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'ddoscoo.DescribeLiveUserDomains'", slog.Any("request", describeDomainsReq), slog.Any("response", describeDomainsResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'ddoscoo.DescribeDomains': %w", err)
	}

	for _, domain := range describeDomainsResp.Body.Domains {
		domains = append(domains, tea.StringValue(domain))
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain string, cloudCertId string) error {
	// 为网站业务转发规则关联 SSL 证书
	// REF: https://help.aliyun.com/zh/anti-ddos/anti-ddos-pro-and-premium/developer-reference/api-ddoscoo-2020-01-01-associatewebcert
	associateWebCertReq := &aliddoscoo.AssociateWebCertRequest{
		Domain:         tea.String(domain),
		CertIdentifier: tea.String(cloudCertId),
	}
	associateWebCertResp, err := d.sdkClient.AssociateWebCertWithContext(ctx, associateWebCertReq, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'ddoscoo.AssociateWebCert'", slog.Any("request", associateWebCertReq), slog.Any("response", associateWebCertResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'ddoscoo.AssociateWebCert': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, accessKeySecret, region string) (*aliddoscoo.Client, error) {
	// 接入点一览 https://api.aliyun.com/product/ddoscoo
	var endpoint string
	switch region {
	case "":
		endpoint = "ddoscoo.cn-hangzhou.aliyuncs.com"
	default:
		endpoint = fmt.Sprintf("ddoscoo.%s.aliyuncs.com", region)
	}

	config := &aliopen.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(endpoint),
	}

	client, err := aliddoscoo.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
