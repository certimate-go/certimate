package qiniucdn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/qiniu-sslcert"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	qiniusdk "github.com/certimate-go/certimate/pkg/sdk3rd/qiniu"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
)

type DeployerConfig struct {
	// 七牛云 AccessKey。
	AccessKey string `json:"accessKey"`
	// 七牛云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 加速域名（支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *qiniusdk.CdnManager
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client := qiniusdk.NewCdnManager(auth.New(config.AccessKey, config.SecretKey))

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
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

	// 获取待部署的域名列表
	var domains []string
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if d.config.Domain == "" {
				return nil, errors.New("config `domain` is required")
			}

			// "*.example.com" → ".example.com"，适配七牛云 CDN 要求的泛域名格式
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
		d.logger.Info("no cdn domains to deploy")
	} else {
		d.logger.Info("found cdn domains to deploy", slog.Any("domains", domains))
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

func (d *Deployer) getAllDomains(ctx context.Context) ([]string, error) {
	domains := make([]string, 0)

	// 查询域名列表
	// REF: https://developer.qiniu.com/fusion/4246/the-domain-name
	getDomainListMarker := ""
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		getDomainListResp, err := d.sdkClient.GetDomainList(ctx, getDomainListMarker, 100)
		d.logger.Debug("sdk request 'cdn.GetDomainList'", slog.String("request.marker", getDomainListMarker), slog.Any("response", getDomainListResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'cdn.GetDomainList': %w", err)
		}

		ignoredStatuses := []string{"frozen", "offlined"}
		for _, domainItem := range getDomainListResp.Domains {
			if lo.Contains(ignoredStatuses, domainItem.OperatingState) {
				continue
			}

			domains = append(domains, domainItem.Name)
		}

		if len(getDomainListResp.Domains) == 0 || getDomainListResp.Marker == "" {
			break
		}

		getDomainListMarker = getDomainListResp.Marker
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain string, cloudCertId string) error {
	// 获取域名信息
	// REF: https://developer.qiniu.com/fusion/4246/the-domain-name
	getDomainInfoResp, err := d.sdkClient.GetDomainInfo(ctx, domain)
	d.logger.Debug("sdk request 'cdn.GetDomainInfo'", slog.String("request.domain", domain), slog.Any("response", getDomainInfoResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'cdn.GetDomainInfo': %w", err)
	}

	// 判断域名是否已启用 HTTPS
	// 如果已启用，修改域名证书；否则，启用 HTTPS
	// REF: https://developer.qiniu.com/fusion/4246/the-domain-name
	if getDomainInfoResp.Https == nil || getDomainInfoResp.Https.CertID == "" {
		enableDomainHttpsResp, err := d.sdkClient.EnableDomainHttps(ctx, domain, cloudCertId, true, true)
		d.logger.Debug("sdk request 'cdn.EnableDomainHttps'", slog.String("request.domain", domain), slog.String("request.certId", cloudCertId), slog.Any("response", enableDomainHttpsResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'cdn.EnableDomainHttps': %w", err)
		}
	} else if getDomainInfoResp.Https.CertID != cloudCertId {
		modifyDomainHttpsConfResp, err := d.sdkClient.ModifyDomainHttpsConf(ctx, domain, cloudCertId, getDomainInfoResp.Https.ForceHttps, getDomainInfoResp.Https.Http2Enable)
		d.logger.Debug("sdk request 'cdn.ModifyDomainHttpsConf'", slog.String("request.domain", domain), slog.String("request.certId", cloudCertId), slog.Any("response", modifyDomainHttpsConfResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'cdn.ModifyDomainHttpsConf': %w", err)
		}
	}

	return nil
}
