package asiaispcdn

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/asiaispcdn"
	asiaispcdnsdk "github.com/certimate-go/certimate/pkg/sdk3rd/asiaispcdn"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// 橙域网络 CDN AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 橙域网络 CDN AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 加速域名（支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *asiaispcdnsdk.Client
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
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
		d.logger.Info("no cdn domains to deploy")
	} else {
		d.logger.Info("found cdn domains to deploy", slog.Any("domains", domains))

		if err := xloop.ForRangeAllWithContext(ctx, domains, func(ctx context.Context, domain string, _ int) error {
			certId, _ := strconv.ParseInt(upres.CertId, 10, 64)
			return d.updateDomainCertificate(ctx, domain, certId)
		}); err != nil {
			return nil, err
		}
	}

	return &DeployResult{}, nil
}

func (d *Deployer) getAllDomains(ctx context.Context) ([]string, error) {
	domains := make([]string, 0)

	// 查询域名信息列表
	domainQueryListReq := &asiaispcdnsdk.DomainQueryListRequest{}
	domainQueryListResp, err := d.sdkClient.DomainQueryListWithContext(ctx, domainQueryListReq)
	d.logger.Debug("sdk request 'domainQueryList'", slog.Any("request", domainQueryListReq), slog.Any("response", domainQueryListResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'domainQueryList': %w", err)
	}

	for _, domainItem := range domainQueryListResp.Data {
		if domainItem.DomainStatus == 0 {
			continue
		}

		domains = append(domains, domainItem.Domain)
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain string, cloudCertId int64) error {
	// 查询域名配置
	domainQueryListReq := &asiaispcdnsdk.DomainQueryListRequest{
		Domain: lo.ToPtr(domain),
	}
	domainQueryListResp, err := d.sdkClient.DomainQueryListWithContext(ctx, domainQueryListReq)
	d.logger.Debug("sdk request 'domainQueryList'", slog.Any("request", domainQueryListReq), slog.Any("response", domainQueryListResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'domainQueryList': %w", err)
	} else {

		domainEntry, _ := lo.Find(domainQueryListResp.Data, func(domainInfo *asiaispcdnsdk.Domain) bool {
			return domainInfo.Domain == domain
		})
		if domainEntry == nil {
			return fmt.Errorf("could not find domain '%s'", domain)
		}

		if domainEntry.CertId == cloudCertId {
			// 已部署过，直接返回
			return nil
		}

		if domainEntry.DomainStatus == 2 {
			if err := d.waitForDomainReady(ctx, domain); err != nil {
				return err
			}
		}
	}

	// 修改域名配置
	domainModifyReq := &asiaispcdnsdk.DomainModifyRequest{
		Domain:   lo.ToPtr(domain),
		Protocol: lo.ToPtr("https"),
		CertId:   lo.ToPtr(cloudCertId),
	}
	domainModifyResp, err := d.sdkClient.DomainModifyWithContext(ctx, domainModifyReq)
	d.logger.Debug("sdk request 'domainModify'", slog.Any("request", domainModifyReq), slog.Any("response", domainModifyResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'domainModify': %w", err)
	}

	return nil
}

func (d *Deployer) waitForDomainReady(ctx context.Context, domain string) error {
	// 查询域名信息，直到操作状态不再为 "2"
	if _, err := xwait.UntilWithContext(ctx, func(_ context.Context, _ int) (bool, error) {
		domainQueryListReq := &asiaispcdnsdk.DomainQueryListRequest{
			Domain: lo.ToPtr(domain),
		}
		domainQueryListResp, err := d.sdkClient.DomainQueryListWithContext(ctx, domainQueryListReq)
		d.logger.Debug("sdk request 'domainQueryList'", slog.Any("request", domainQueryListReq), slog.Any("response", domainQueryListResp))
		if err != nil {
			return false, fmt.Errorf("failed to execute sdk request 'domainQueryList': %w", err)
		}

		domainEntry, _ := lo.Find(domainQueryListResp.Data, func(domainInfo *asiaispcdnsdk.Domain) bool {
			return domainInfo.Domain == domain
		})
		if domainEntry == nil {
			return false, fmt.Errorf("could not find domain '%s'", domain)
		}

		if domainEntry.DomainStatus != 2 {
			return true, nil
		}

		d.logger.Info("waiting for domain's status to not be 'Operating' ...")
		return false, nil
	}, 10*time.Second); err != nil {
		return err
	}

	return nil
}

func createSDKClient(accessKeyId, accessKeySecret string) (*asiaispcdnsdk.Client, error) {
	client, err := asiaispcdnsdk.NewClient(
		asiaispcdnsdk.WithAkSk(accessKeyId, accessKeySecret),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
