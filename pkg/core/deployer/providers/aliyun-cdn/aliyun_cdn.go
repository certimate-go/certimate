package aliyuncdn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	alicdn "github.com/alibabacloud-go/cdn-20180510/v9/client"
	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aliyun-cas"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-cdn/internal"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
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

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret)
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
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
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
				return nil, errors.New("config `domain` is required")
			}

			// "*.example.com" → ".example.com"，适配阿里云 CDN 要求的泛域名格式
			domain := strings.TrimPrefix(d.config.Domain, "*")
			domains = []string{domain}
		}

	case DOMAIN_MATCH_PATTERN_WILDCARD:
		{
			if d.config.Domain == "" {
				return nil, errors.New("config `domain` is required")
			}

			if strings.HasPrefix(d.config.Domain, "*.") {
				domainCandidates, err := d.getAllDomains(ctx)
				if err != nil {
					return nil, err
				}

				domains = lo.Filter(domainCandidates, func(domain string, _ int) bool {
					return xcerthostname.IsMatch(d.config.Domain, domain) ||
						strings.TrimPrefix(d.config.Domain, "*") == strings.TrimPrefix(domain, "*")
				})
				if len(domains) == 0 {
					return nil, errors.New("could not find any domains matched by wildcard")
				}
			} else {
				domains = []string{d.config.Domain}
			}
		}

	case DOMAIN_MATCH_PATTERN_CERTSAN:
		{
			certX509, err := xcert.ParseCertificateFromPEM(certPEM)
			if err != nil {
				return nil, err
			}

			domainCandidates, err := d.getAllDomains(ctx)
			if err != nil {
				return nil, err
			}

			domains = lo.Filter(domainCandidates, func(domain string, _ int) bool {
				return certX509.VerifyHostname(domain) == nil ||
					strings.TrimPrefix(d.config.Domain, "*") == strings.TrimPrefix(domain, "*")
			})
			if len(domains) == 0 {
				return nil, errors.New("could not find any domains matched by certificate")
			}
		}

	default:
		return nil, fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 遍历更新域名证书
	if len(domains) == 0 {
		d.logger.Info("no cdn domains to deploy")
	} else {
		d.logger.Info("found cdn domains to deploy", slog.Any("domains", domains))
		var errs []error

		certIdentifier := upres.ExtendedData["CertIdentifier"].(string)
		certIdentifierSeps := strings.SplitN(certIdentifier, "-", 2)
		if len(certIdentifierSeps) != 2 {
			return nil, fmt.Errorf("received invalid certificate identifier: '%s'", certIdentifier)
		}

		certId, _ := strconv.ParseInt(certIdentifierSeps[0], 10, 64)
		certRegion := certIdentifierSeps[1]
		for _, domain := range domains {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				if err := d.updateDomainCertificate(ctx, domain, certId, certRegion); err != nil {
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

func (d *Deployer) getAllDomains(ctx context.Context) ([]string, error) {
	domains := make([]string, 0)

	// 查询域名列表
	// REF: https://help.aliyun.com/zh/cdn/developer-reference/api-cdn-2018-05-10-describeuserdomains
	describeUserDomainsPageNumber := 1
	describeUserDomainsPageSize := 500
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		describeUserDomainsReq := &alicdn.DescribeUserDomainsRequest{
			ResourceGroupId: lo.EmptyableToPtr(d.config.ResourceGroupId),
			PageNumber:      tea.Int32(int32(describeUserDomainsPageNumber)),
			PageSize:        tea.Int32(int32(describeUserDomainsPageSize)),
		}
		describeUserDomainsResp, err := d.sdkClient.DescribeUserDomainsWithContext(ctx, describeUserDomainsReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'cdn.DescribeUserDomains'", slog.Any("request", describeUserDomainsReq), slog.Any("response", describeUserDomainsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'cdn.DescribeUserDomains': %w", err)
		}

		if describeUserDomainsResp.Body == nil || describeUserDomainsResp.Body.Domains == nil {
			break
		}

		ignoredStatuses := []string{"offline", "checking", "check_failed", "stopping", "deleting"}
		for _, domainItem := range describeUserDomainsResp.Body.Domains.PageData {
			if lo.Contains(ignoredStatuses, tea.StringValue(domainItem.DomainStatus)) {
				continue
			}

			domains = append(domains, tea.StringValue(domainItem.DomainName))
		}

		if len(describeUserDomainsResp.Body.Domains.PageData) < describeUserDomainsPageSize {
			break
		}

		describeUserDomainsPageNumber++
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain string, cloudCertId int64, certRegion string) error {
	// 设置 CDN 域名域名证书
	// REF: https://help.aliyun.com/zh/cdn/developer-reference/api-cdn-2018-05-10-setcdndomainsslcertificate
	setCdnDomainSSLCertificateReq := &alicdn.SetCdnDomainSSLCertificateRequest{
		DomainName:  tea.String(domain),
		CertType:    tea.String("cas"),
		CertId:      tea.Int64(cloudCertId),
		CertRegion:  tea.String(certRegion),
		SSLProtocol: tea.String("on"),
	}
	setCdnDomainSSLCertificateResp, err := d.sdkClient.SetCdnDomainSSLCertificateWithContext(ctx, setCdnDomainSSLCertificateReq, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'cdn.SetCdnDomainSSLCertificate'", slog.Any("request", setCdnDomainSSLCertificateReq), slog.Any("response", setCdnDomainSSLCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'cdn.SetCdnDomainSSLCertificate': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, accessKeySecret string) (*internal.CdnClient, error) {
	config := &aliopen.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String("cdn.aliyuncs.com"),
	}

	client, err := internal.NewCdnClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
