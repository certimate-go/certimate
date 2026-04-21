package jdcloudlive

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	jdcore "github.com/jdcloud-api/jdcloud-sdk-go/core"
	jdlive "github.com/jdcloud-api/jdcloud-sdk-go/services/live/apis"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/jdcloud-live/internal"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
)

type DeployerConfig struct {
	// 京东云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 京东云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 直播流域名（不支持泛域名）。
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

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret)
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

			domains = []string{d.config.Domain}
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

	// 查询域名列表
	// REF: https://docs.jdcloud.com/cn/live-video/api/describelivedomains
	describeLiveDomainsPageNumber := 1
	describeLiveDomainsPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		describeLiveDomainsReq := jdlive.NewDescribeLiveDomainsRequestWithoutParam()
		describeLiveDomainsReq.SetPageNum(describeLiveDomainsPageNumber)
		describeLiveDomainsReq.SetPageSize(describeLiveDomainsPageSize)
		describeLiveDomainsResp, err := d.sdkClient.DescribeLiveDomains(describeLiveDomainsReq)
		d.logger.Debug("sdk request 'live.DescribeLiveDomainsRequest'", slog.Any("request", describeLiveDomainsReq), slog.Any("response", describeLiveDomainsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'live.DescribeLiveDomainsRequest': %w", err)
		}

		ignoredStatuses := []string{"offline", "checking", "check_failed"}
		for _, domainItem := range describeLiveDomainsResp.Result.DomainDetails {
			for _, playDomainItem := range domainItem.PlayDomains {
				if lo.Contains(ignoredStatuses, playDomainItem.DomainStatus) {
					continue
				}

				domains = append(domains, playDomainItem.PlayDomain)
			}
		}

		if len(describeLiveDomainsResp.Result.DomainDetails) < describeLiveDomainsPageSize {
			break
		}

		describeLiveDomainsPageNumber++
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain string, certPEM, privkeyPEM string) error {
	// 设置直播证书
	// REF: https://docs.jdcloud.com/cn/live-video/api/setlivedomaincertificate
	setLiveDomainCertificateReq := jdlive.NewSetLiveDomainCertificateRequestWithoutParam()
	setLiveDomainCertificateReq.SetPlayDomain(domain)
	setLiveDomainCertificateReq.SetCertStatus("on")
	setLiveDomainCertificateReq.SetCert(certPEM)
	setLiveDomainCertificateReq.SetKey(privkeyPEM)
	setLiveDomainCertificateResp, err := d.sdkClient.SetLiveDomainCertificate(setLiveDomainCertificateReq)
	d.logger.Debug("sdk request 'live.SetLiveDomainCertificate'", slog.Any("request", setLiveDomainCertificateReq), slog.Any("response", setLiveDomainCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'live.SetLiveDomainCertificate': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, accessKeySecret string) (*internal.LiveClient, error) {
	clientCredentials := jdcore.NewCredentials(accessKeyId, accessKeySecret)
	client := internal.NewLiveClient(clientCredentials)
	return client, nil
}
