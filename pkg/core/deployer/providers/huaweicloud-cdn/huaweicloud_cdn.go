package huaweicloudcdn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	hccdn "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdn/v2"
	hccdnmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdn/v2/model"
	hccdnregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdn/v2/region"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/huaweicloud-scm"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/huaweicloud-cdn/internal"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
)

type DeployerConfig struct {
	// 华为云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 华为云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 华为云企业项目 ID。
	EnterpriseProjectId string `json:"enterpriseProjectId,omitempty"`
	// 华为云区域。
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

	client, err := createSDKClient(
		config.AccessKeyId,
		config.SecretAccessKey,
		config.Region,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		AccessKeyId:         config.AccessKeyId,
		SecretAccessKey:     config.SecretAccessKey,
		EnterpriseProjectId: config.EnterpriseProjectId,
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
		d.logger.Info("no cdn domains to deploy")
	} else {
		d.logger.Info("found cdn domains to deploy", slog.Any("domains", domains))
		var errs []error

		const MAX_DOMAIN_PER_REQUEST = 50
		domainChunks := lo.Chunk(domains, MAX_DOMAIN_PER_REQUEST)
		for _, domains := range domainChunks {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				if err := d.updateDomainsCertificate(ctx, domains, upres.CertId, upres.CertName); err != nil {
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
	// REF: https://support.huaweicloud.com/api-cdn/ListDomains.html
	listDomainsPageNumber := 1
	listDomainsPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listDomainsReq := &hccdnmodel.ListDomainsRequest{
			EnterpriseProjectId: lo.EmptyableToPtr(d.config.EnterpriseProjectId),
			PageNumber:          lo.ToPtr(int32(listDomainsPageNumber)),
			PageSize:            lo.ToPtr(int32(listDomainsPageSize)),
		}
		listDomainsResp, err := d.sdkClient.ListDomains(listDomainsReq)
		d.logger.Debug("sdk request 'cdn.ListDomains'", slog.Any("request", listDomainsReq), slog.Any("response", listDomainsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'cdn.ListDomains': %w", err)
		}

		if listDomainsResp.Domains == nil {
			break
		}

		ignoredStatuses := []string{"offline", "checking", "check_failed", "deleting"}
		for _, domainItem := range *listDomainsResp.Domains {
			if lo.Contains(ignoredStatuses, lo.FromPtr(domainItem.DomainStatus)) {
				continue
			}

			domains = append(domains, lo.FromPtr(domainItem.DomainName))
		}

		if len(*listDomainsResp.Domains) < listDomainsPageSize {
			break
		}

		listDomainsPageNumber++
	}

	return domains, nil
}

func (d *Deployer) updateDomainsCertificate(ctx context.Context, domains []string, cloudCertId, cloudCertName string) error {
	// 更新加速域名配置
	// REF: https://support.huaweicloud.com/api-cdn/UpdateDomainMultiCertificates.html
	// REF: https://support.huaweicloud.com/usermanual-cdn/cdn_01_0306.html
	updateDomainMultiCertificatesReq := &hccdnmodel.UpdateDomainMultiCertificatesRequest{
		EnterpriseProjectId: lo.EmptyableToPtr(d.config.EnterpriseProjectId),
		Body: &hccdnmodel.UpdateDomainMultiCertificatesRequestBody{
			Https: &hccdnmodel.UpdateDomainMultiCertificatesRequestBodyContent{
				DomainName:       strings.Join(domains, ","),
				HttpsSwitch:      1,
				CertificateType:  lo.ToPtr(int32(2)),
				ScmCertificateId: lo.ToPtr(cloudCertId),
				CertName:         lo.ToPtr(cloudCertName),
			},
		},
	}
	updateDomainMultiCertificatesResp, err := d.sdkClient.UpdateDomainMultiCertificates(updateDomainMultiCertificatesReq)
	d.logger.Debug("sdk request 'cdn.UpdateDomainMultiCertificates'", slog.Any("request", updateDomainMultiCertificatesReq), slog.Any("response", updateDomainMultiCertificatesResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'cdn.UpdateDomainMultiCertificates': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*internal.CdnClient, error) {
	if region == "" {
		region = "cn-north-1" // CDN 服务默认区域：华北一北京
	}

	auth, err := global.NewCredentialsBuilder().
		WithAk(accessKeyId).
		WithSk(secretAccessKey).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	hcRegion, err := hccdnregion.SafeValueOf(region)
	if err != nil {
		return nil, err
	}

	hcClient, err := hccdn.CdnClientBuilder().
		WithRegion(hcRegion).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	client := internal.NewCdnClient(hcClient)
	return client, nil
}
