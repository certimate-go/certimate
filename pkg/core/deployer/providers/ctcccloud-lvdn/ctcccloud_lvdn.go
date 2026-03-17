package ctcccloudlvdn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/ctcccloud-lvdn"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	ctyunlvdn "github.com/certimate-go/certimate/pkg/sdk3rd/ctyun/lvdn"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type DeployerConfig struct {
	// 天翼云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 天翼云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 加速域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *ctyunlvdn.Client
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
		d.logger.Info("no lvdn domains to deploy")
	} else {
		d.logger.Info("found lvdn domains to deploy", slog.Any("domains", domains))
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
	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=125&api=11559&data=183&isNormal=1&vid=261
	queryDomainsPage := 1
	queryDomainsPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		queryDomainListReq := &ctyunlvdn.QueryDomainListRequest{
			Page:        lo.ToPtr(int32(queryDomainsPage)),
			PageSize:    lo.ToPtr(int32(queryDomainsPageSize)),
			ProductCode: lo.ToPtr("005"),
		}
		queryDomainListResp, err := d.sdkClient.QueryDomainListWithContext(ctx, queryDomainListReq)
		d.logger.Debug("sdk request 'cdn.QueryDomainList'", slog.Any("request", queryDomainListReq), slog.Any("response", queryDomainListResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'cdn.QueryDomainList': %w", err)
		}

		if queryDomainListResp.ReturnObj == nil {
			break
		}

		ignoredStatuses := []int32{1, 5, 6, 7, 8, 9, 11, 12}
		for _, domainItem := range queryDomainListResp.ReturnObj.Results {
			if lo.Contains(ignoredStatuses, domainItem.Status) {
				continue
			}

			domains = append(domains, domainItem.Domain)
		}

		if len(queryDomainListResp.ReturnObj.Results) < queryDomainsPageSize {
			break
		}

		queryDomainsPage++
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain string, cloudCertName string) error {
	// 查询域名配置信息
	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=125&api=11473&data=183&isNormal=1&vid=261
	queryDomainDetailReq := &ctyunlvdn.QueryDomainDetailRequest{
		Domain:      lo.ToPtr(domain),
		ProductCode: lo.ToPtr("005"),
	}
	queryDomainDetailResp, err := d.sdkClient.QueryDomainDetailWithContext(ctx, queryDomainDetailReq)
	d.logger.Debug("sdk request 'lvdn.QueryDomainDetail'", slog.Any("request", queryDomainDetailReq), slog.Any("response", queryDomainDetailResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'lvdn.QueryDomainDetail': %w", err)
	}

	// 修改域名配置
	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=108&api=11308&data=161&isNormal=1&vid=154
	updateDomainReq := &ctyunlvdn.UpdateDomainRequest{
		Domain:      lo.ToPtr(domain),
		ProductCode: lo.ToPtr("005"),
		HttpsSwitch: lo.ToPtr(int32(1)),
		CertName:    lo.ToPtr(cloudCertName),
	}
	updateDomainResp, err := d.sdkClient.UpdateDomainWithContext(ctx, updateDomainReq)
	d.logger.Debug("sdk request 'lvdn.UpdateDomain'", slog.Any("request", updateDomainReq), slog.Any("response", updateDomainResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'lvdn.UpdateDomain': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, secretAccessKey string) (*ctyunlvdn.Client, error) {
	return ctyunlvdn.NewClient(accessKeyId, secretAccessKey)
}
