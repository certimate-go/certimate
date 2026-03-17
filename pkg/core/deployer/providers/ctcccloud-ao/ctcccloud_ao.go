package ctcccloudao

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/ctcccloud-ao"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	ctyunao "github.com/certimate-go/certimate/pkg/sdk3rd/ctyun/ao"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
)

type DeployerConfig struct {
	// 天翼云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 天翼云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 加速域名（支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *ctyunao.Client
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		SecretAccessKey: config.SecretAccessKey,
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
		d.logger.Info("no accessone domains to deploy")
	} else {
		d.logger.Info("found accessone domains to deploy", slog.Any("domains", domains))
		var errs []error

		for _, domain := range domains {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				if err := d.updateDomainCertificate(ctx, domain, upres.CertName); err != nil {
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
	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=113&api=13816&data=174&isNormal=1&vid=167
	queryDomainsPage := 1
	queryDomainsPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		queryDomainsReq := &ctyunao.QueryDomainsRequest{
			Page:        lo.ToPtr(int32(queryDomainsPage)),
			PageSize:    lo.ToPtr(int32(queryDomainsPageSize)),
			ProductCode: lo.ToPtr("020"),
		}
		queryDomainsResp, err := d.sdkClient.QueryDomainsWithContext(ctx, queryDomainsReq)
		d.logger.Debug("sdk request 'cdn.QueryDomains'", slog.Any("request", queryDomainsReq), slog.Any("response", queryDomainsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'cdn.QueryDomains': %w", err)
		}

		if queryDomainsResp.ReturnObj == nil {
			break
		}

		ignoredStatuses := []int32{1, 5, 6, 7, 8, 9, 11, 12}
		for _, domainItem := range queryDomainsResp.ReturnObj.Results {
			if lo.Contains(ignoredStatuses, domainItem.Status) {
				continue
			}

			domains = append(domains, domainItem.Domain)
		}

		if len(queryDomainsResp.ReturnObj.Results) < queryDomainsPageSize {
			break
		}

		queryDomainsPage++
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain string, cloudCertName string) error {
	// 域名基础及加速配置查询
	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=113&api=13412&data=174&isNormal=1&vid=167
	getDomainConfigReq := &ctyunao.GetDomainConfigRequest{
		Domain:      lo.ToPtr(domain),
		ProductCode: lo.ToPtr("020"),
	}
	getDomainConfigResp, err := d.sdkClient.GetDomainConfigWithContext(ctx, getDomainConfigReq)
	d.logger.Debug("sdk request 'cdn.GetDomainConfig'", slog.Any("request", getDomainConfigReq), slog.Any("response", getDomainConfigResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'cdn.GetDomainConfig': %w", err)
	}

	// 域名基础及加速配置修改
	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=113&api=13413&data=174&isNormal=1&vid=167
	modifyDomainConfigReq := &ctyunao.ModifyDomainConfigRequest{
		Domain:      lo.ToPtr(domain),
		ProductCode: lo.ToPtr(getDomainConfigResp.ReturnObj.ProductCode),
		Origin: lo.Map(getDomainConfigResp.ReturnObj.Origin, func(item *ctyunao.DomainOriginConfigWithWeight, _ int) *ctyunao.DomainOriginConfig {
			weight := item.Weight
			if weight == 0 {
				weight = 1
			}
			return &ctyunao.DomainOriginConfig{
				Origin: item.Origin,
				Role:   item.Role,
				Weight: strconv.Itoa(int(weight)),
			}
		}),
		HttpsStatus: lo.ToPtr("on"),
		CertName:    lo.ToPtr(cloudCertName),
	}
	modifyDomainConfigResp, err := d.sdkClient.ModifyDomainConfigWithContext(ctx, modifyDomainConfigReq)
	d.logger.Debug("sdk request 'cdn.ModifyDomainConfig'", slog.Any("request", modifyDomainConfigReq), slog.Any("response", modifyDomainConfigResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'cdn.ModifyDomainConfig': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, secretAccessKey string) (*ctyunao.Client, error) {
	return ctyunao.NewClient(accessKeyId, secretAccessKey)
}
