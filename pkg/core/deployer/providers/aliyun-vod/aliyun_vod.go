package aliyunvod

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	alivod "github.com/alibabacloud-go/vod-20170321/v4/client"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aliyun-cas"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-vod/internal"
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
	// 点播加速域名（支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *internal.VodClient
	sdkCertmgr certmgr.Provider
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

			domains = []string{d.config.Domain}
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
			certX509, err := xcert.ParseCertificateFromPEM(certPEM)
			if err != nil {
				return nil, err
			}

			domainCandidates, err := d.getAllDomains(ctx)
			if err != nil {
				return nil, err
			}

			domains = lo.Filter(domainCandidates, func(domain string, _ int) bool {
				return certX509.VerifyHostname(domain) == nil
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
		d.logger.Info("no vod domains to deploy")
	} else {
		d.logger.Info("found vod domains to deploy", slog.Any("domains", domains))
		var errs []error

		for _, domain := range domains {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				certId, _ := strconv.ParseInt(upres.CertId, 10, 64)
				certName := upres.CertName
				if err := d.updateDomainCertificate(ctx, domain, certId, certName); err != nil {
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

	// 查询加速域名列表
	// REF: https://help.aliyun.com/zh/live/developer-reference/api-live-2016-11-01-describeliveuserdomains
	describeVodUserDomainsPageNumber := 1
	describeVodUserDomainsPageSize := 50
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		describeVodUserDomainsReq := &alivod.DescribeVodUserDomainsRequest{
			DomainStatus: tea.String("online"),
			PageNumber:   tea.Int32(int32(describeVodUserDomainsPageNumber)),
			PageSize:     tea.Int32(int32(describeVodUserDomainsPageSize)),
		}
		describeVodUserDomainsResp, err := d.sdkClient.DescribeVodUserDomainsWithContext(ctx, describeVodUserDomainsReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'vod.DescribeVodUserDomains'", slog.Any("request", describeVodUserDomainsReq), slog.Any("response", describeVodUserDomainsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'vod.DescribeLiveUserDomains': %w", err)
		}

		if describeVodUserDomainsResp.Body == nil || describeVodUserDomainsResp.Body.Domains == nil {
			break
		}

		for _, domainItem := range describeVodUserDomainsResp.Body.Domains.PageData {
			domains = append(domains, tea.StringValue(domainItem.DomainName))
		}

		if len(describeVodUserDomainsResp.Body.Domains.PageData) < describeVodUserDomainsPageSize {
			break
		}

		describeVodUserDomainsPageNumber++
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain string, cloudCertId int64, cloudCertName string) error {
	// 设置域名证书
	// REF: https://help.aliyun.com/zh/vod/developer-reference/api-vod-2017-03-21-setvoddomainsslcertificate
	setVodDomainSSLCertificateReq := &alivod.SetVodDomainSSLCertificateRequest{
		DomainName: tea.String(domain),
		CertType:   tea.String("cas"),
		CertId:     tea.Int64(cloudCertId),
		CertName:   tea.String(cloudCertName),
		CertRegion: lo.
			If(d.config.Region == "" || strings.HasPrefix(d.config.Region, "cn-"), tea.String("cn-hangzhou")).
			Else(tea.String("ap-southeast-1")),
		SSLProtocol: tea.String("on"),
	}
	setVodDomainSSLCertificateResp, err := d.sdkClient.SetVodDomainSSLCertificateWithContext(ctx, setVodDomainSSLCertificateReq, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'live.SetVodDomainSSLCertificate'", slog.Any("request", setVodDomainSSLCertificateReq), slog.Any("response", setVodDomainSSLCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'live.SetVodDomainSSLCertificate': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, accessKeySecret, region string) (*internal.VodClient, error) {
	// 接入点一览 https://api.aliyun.com/product/vod
	var endpoint string
	switch region {
	case "":
		endpoint = "vod.cn-hangzhou.aliyuncs.com"
	default:
		endpoint = fmt.Sprintf("vod.%s.aliyuncs.com", region)
	}

	config := &aliopen.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(endpoint),
	}

	client, err := internal.NewVodClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
