package volcenginevod

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/samber/lo"
	ve "github.com/volcengine/volcengine-go-sdk/volcengine"
	vesession "github.com/volcengine/volcengine-go-sdk/volcengine/session"

	vevod "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/volcengine/volcengine-go-sdk/service/vod20260101"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/volcengine-certcenter"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// 火山引擎 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 火山引擎 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 火山引擎项目名称。
	ProjectName string `json:"projectName,omitempty"`
	// 火山引擎地域。
	Region string `json:"region"`
	// 点播空间名称。
	SpaceName string `json:"spaceName"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 点播域名类型。
	DomainType string `json:"domainType"`
	// 点播加速域名（支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *vevod.VOD20260101
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		SecretAccessKey: config.SecretAccessKey,
		ProjectName:     config.ProjectName,
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

	// 获取待部署的域名
	domains := make([]string, 0)
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if d.config.Domain == "" {
				return nil, fmt.Errorf("config `domain` is required")
			}

			domains = append(domains, d.config.Domain)
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
				domains = append(domains, d.config.Domain)
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
		d.logger.Info("no vod domains to deploy")
	} else {
		d.logger.Info("found vod domains to deploy", slog.Any("domains", domains))

		if err := xloop.ForRangeAllWithContext(ctx, domains, func(ctx context.Context, domain string, _ int) error {
			return d.updateDomainCertificate(ctx, domain, upres.CertId)
		}); err != nil {
			return nil, err
		}
	}

	return &DeployResult{}, nil
}

func (d *Deployer) getAllDomains(ctx context.Context) ([]string, error) {
	domains := make([]string, 0)

	// 获取域名列表
	// REF: https://www.volcengine.com/docs/4/2389927
	listVodDomainPageNum := 1
	listVodDomainPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listVodDomainReq := &vevod.ListVodDomainsInput{
			SpaceName:  ve.String(d.config.SpaceName),
			DomainType: ve.String(convertDomainType2CloudDomainType(d.config.DomainType)),
			ListCdnDomainsParam: &vevod.ListCdnDomainsParamForListVodDomainsInput{
				PageNum:  ve.Int64(int64(listVodDomainPageNum)),
				PageSize: ve.Int64(int64(listVodDomainPageSize)),
			},
		}
		listVodDomainResp, err := d.sdkClient.ListVodDomainsWithContext(ctx, listVodDomainReq)
		d.logger.Debug("sdk request 'vod.ListVodDomain'", slog.Any("request", listVodDomainReq), slog.Any("response", listVodDomainResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'vod.ListVodDomain': %w", err)
		}

		if listVodDomainResp.VodInfo == nil {
			break
		}

		for _, domainItem := range listVodDomainResp.VodInfo.Domains {
			domains = append(domains, ve.StringValue(domainItem.Domain))
		}

		if len(listVodDomainResp.VodInfo.Domains) < listVodDomainPageSize {
			break
		}

		listVodDomainPageNum++
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain string, cloudCertId string) error {
	// 更新域名配置
	// REF: https://www.volcengine.com/docs/4/2389907
	updateVodDomainConfigReq := &vevod.UpdateVodDomainConfigInput{
		SpaceName:  ve.String(d.config.SpaceName),
		DomainType: ve.String(convertDomainType2CloudDomainType(d.config.DomainType)),
		UpdateCdnConfigParam: &vevod.UpdateCdnConfigParamForUpdateVodDomainConfigInput{
			Domain: ve.String(domain),
			HTTPS: &vevod.HTTPSForUpdateVodDomainConfigInput{
				Switch: ve.Bool(true),
				CertInfo: &vevod.CertInfoForUpdateVodDomainConfigInput{
					CertId: ve.String(cloudCertId),
				},
			},
		},
	}
	updateVodDomainConfigResp, err := d.sdkClient.UpdateVodDomainConfigWithContext(ctx, updateVodDomainConfigReq)
	d.logger.Debug("sdk request 'vod.UpdateVodDomainConfig'", slog.Any("request", updateVodDomainConfigReq), slog.Any("response", updateVodDomainConfigResp))
	if err != nil {
		return err
	}

	return nil
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*vevod.VOD20260101, error) {
	if region == "" {
		region = "cn-north-1" // VOD 服务默认区域：华北
	}

	config := ve.NewConfig().
		WithAkSk(accessKeyId, secretAccessKey).
		WithRegion(region)

	session, err := vesession.NewSession(config)
	if err != nil {
		return nil, err
	}

	client := vevod.New(session)
	return client, nil
}

func convertDomainType2CloudDomainType(domainType string) string {
	switch domainType {
	case DOMAIN_TYPE_PLAY:
		return "vod_play"
	case DOMAIN_TYPE_IMAGE:
		return "vod_image"
	case DOMAIN_TYPE_THIRD:
		return "third"
	default:
		return domainType
	}
}
