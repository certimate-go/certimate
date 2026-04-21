package aliyunlive

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	alilive "github.com/alibabacloud-go/live-20161101/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-live/internal"
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
	// 直播流域名（支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *internal.LiveClient
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return &Deployer{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
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
	// 获取待部署的域名列表
	var domains []string
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if d.config.Domain == "" {
				return nil, errors.New("config `domain` is required")
			}

			// "*.example.com" → ".example.com"，适配阿里云 Live 要求的泛域名格式
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
					return xcerthostname.IsMatch(d.config.Domain, domain)
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
			domainCandidates, err := d.getAllDomains(ctx)
			if err != nil {
				return nil, err
			}

			domains = lo.Filter(domainCandidates, func(domain string, _ int) bool {
				return xcerthostname.IsMatchByCertificatePEM(certPEM, domain)
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
		d.logger.Info("no live domains to deploy")
	} else {
		d.logger.Info("found live domains to deploy", slog.Any("domains", domains))
		var errs []error

		for _, domain := range domains {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				if err := d.updateDomainCertificate(ctx, domain, certPEM, privkeyPEM); err != nil {
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

	// 查询用户名下所有的直播域名
	// REF: https://help.aliyun.com/zh/live/developer-reference/api-live-2016-11-01-describeliveuserdomains
	describeUserLiveDomainsPageNumber := 1
	describeUserLiveDomainsPageSize := 50
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		describeUserLiveDomainsReq := &alilive.DescribeLiveUserDomainsRequest{
			ResourceGroupId: lo.EmptyableToPtr(d.config.ResourceGroupId),
			RegionName:      tea.String(d.config.Region),
			DomainStatus:    tea.String("online"),
			PageNumber:      tea.Int32(int32(describeUserLiveDomainsPageNumber)),
			PageSize:        tea.Int32(int32(describeUserLiveDomainsPageSize)),
		}
		describeUserLiveDomainsResp, err := d.sdkClient.DescribeLiveUserDomainsWithContext(ctx, describeUserLiveDomainsReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'live.DescribeLiveUserDomains'", slog.Any("request", describeUserLiveDomainsReq), slog.Any("response", describeUserLiveDomainsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'live.DescribeLiveUserDomains': %w", err)
		}

		if describeUserLiveDomainsResp.Body == nil || describeUserLiveDomainsResp.Body.Domains == nil {
			break
		}

		for _, domainItem := range describeUserLiveDomainsResp.Body.Domains.PageData {
			domains = append(domains, tea.StringValue(domainItem.DomainName))
		}

		if len(describeUserLiveDomainsResp.Body.Domains.PageData) < describeUserLiveDomainsPageSize {
			break
		}

		describeUserLiveDomainsPageNumber++
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain string, certPEM, privkeyPEM string) error {
	// 设置域名证书
	// REF: https://help.aliyun.com/zh/live/developer-reference/api-live-2016-11-01-setlivedomaincertificate
	setLiveDomainSSLCertificateReq := &alilive.SetLiveDomainCertificateRequest{
		DomainName:  tea.String(domain),
		CertName:    tea.String(fmt.Sprintf("certimate-%d", time.Now().UnixMilli())),
		CertType:    tea.String("upload"),
		SSLProtocol: tea.String("on"),
		SSLPub:      tea.String(certPEM),
		SSLPri:      tea.String(privkeyPEM),
	}
	setLiveDomainSSLCertificateResp, err := d.sdkClient.SetLiveDomainCertificateWithContext(ctx, setLiveDomainSSLCertificateReq, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'live.SetLiveDomainCertificate'", slog.Any("request", setLiveDomainSSLCertificateReq), slog.Any("response", setLiveDomainSSLCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'live.SetLiveDomainCertificate': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, accessKeySecret, region string) (*internal.LiveClient, error) {
	// 接入点一览 https://api.aliyun.com/product/live
	var endpoint string
	switch region {
	case "",
		"cn-qingdao",
		"cn-beijing",
		"cn-shanghai",
		"cn-shenzhen",
		"ap-northeast-1",
		"ap-southeast-5",
		"me-central-1":
		endpoint = "live.aliyuncs.com"
	default:
		endpoint = fmt.Sprintf("live.%s.aliyuncs.com", region)
	}

	config := &aliopen.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(endpoint),
	}

	client, err := internal.NewLiveClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
