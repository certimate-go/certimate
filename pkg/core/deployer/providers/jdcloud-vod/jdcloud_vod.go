package jdcloudvod

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	jdcore "github.com/jdcloud-api/jdcloud-sdk-go/core"
	jdvod "github.com/jdcloud-api/jdcloud-sdk-go/services/vod/apis"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/jdcloud-vod/internal"
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
	// 点播加速域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *internal.VodClient
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
		d.logger.Info("no vod domains to deploy")
	} else {
		d.logger.Info("found vod domains to deploy", slog.Any("domains", domains))
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
	// REF: https://docs.jdcloud.com/cn/video-on-demand/api/listdomains
	listDomainsPageNumber := 1
	listDomainsPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listDomainsReq := jdvod.NewListDomainsRequestWithoutParam()
		listDomainsReq.SetPageNumber(listDomainsPageNumber)
		listDomainsReq.SetPageSize(listDomainsPageSize)
		listDomainsResp, err := d.sdkClient.ListDomains(listDomainsReq)
		d.logger.Debug("sdk request 'vod.ListDomains'", slog.Any("request", listDomainsReq), slog.Any("response", listDomainsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'vod.ListDomains': %w", err)
		}

		ignoredStatuses := []string{"init", "stopped"}
		for _, domainItem := range listDomainsResp.Result.Content {
			if lo.Contains(ignoredStatuses, domainItem.Status) {
				continue
			}

			domains = append(domains, domainItem.Name)
		}

		if len(listDomainsResp.Result.Content) < listDomainsPageSize {
			break
		}

		listDomainsPageNumber++
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain string, certPEM, privkeyPEM string) error {
	// 获取域名 ID
	domainId, err := d.findDomainIdByDomain(ctx, domain)
	if err != nil {
		return err
	}

	// 查询域名 SSL 配置
	// REF: https://docs.jdcloud.com/cn/video-on-demand/api/gethttpssl
	getHttpSslReq := jdvod.NewGetHttpSslRequestWithoutParam()
	getHttpSslReq.SetDomainId(domainId)
	getHttpSslResp, err := d.sdkClient.GetHttpSsl(getHttpSslReq)
	d.logger.Debug("sdk request 'vod.GetHttpSsl'", slog.Any("request", getHttpSslReq), slog.Any("response", getHttpSslResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'vod.GetHttpSsl': %w", err)
	}

	// 设置域名 SSL 配置
	// REF: https://docs.jdcloud.com/cn/video-on-demand/api/sethttpssl
	setHttpSslReq := jdvod.NewSetHttpSslRequestWithoutParam()
	setHttpSslReq.SetDomainId(domainId)
	setHttpSslReq.SetTitle(fmt.Sprintf("certimate-%d", time.Now().UnixMilli()))
	setHttpSslReq.SetSslCert(certPEM)
	setHttpSslReq.SetSslKey(privkeyPEM)
	setHttpSslReq.SetSource("default")
	setHttpSslReq.SetJumpType(getHttpSslResp.Result.JumpType)
	setHttpSslReq.SetEnabled(true)
	setHttpSslResp, err := d.sdkClient.SetHttpSsl(setHttpSslReq)
	d.logger.Debug("sdk request 'vod.SetHttpSsl'", slog.Any("request", setHttpSslReq), slog.Any("response", setHttpSslResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'vod.SetHttpSsl': %w", err)
	}

	return nil
}

func (d *Deployer) findDomainIdByDomain(ctx context.Context, domain string) (int, error) {
	// 查询域名列表
	// REF: https://docs.jdcloud.com/cn/video-on-demand/api/listdomains
	listDomainsPageNumber := 1
	listDomainsPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
		}

		listDomainsReq := jdvod.NewListDomainsRequestWithoutParam()
		listDomainsReq.SetPageNumber(listDomainsPageNumber)
		listDomainsReq.SetPageSize(listDomainsPageSize)
		listDomainsResp, err := d.sdkClient.ListDomains(listDomainsReq)
		d.logger.Debug("sdk request 'vod.ListDomains'", slog.Any("request", listDomainsReq), slog.Any("response", listDomainsResp))
		if err != nil {
			return 0, fmt.Errorf("failed to execute sdk request 'vod.ListDomains': %w", err)
		}

		for _, domainItem := range listDomainsResp.Result.Content {
			if domainItem.Name == domain {
				domainId, _ := strconv.Atoi(domainItem.Id)
				return domainId, nil
			}
		}

		if len(listDomainsResp.Result.Content) < listDomainsPageSize {
			break
		}

		listDomainsPageNumber++
	}

	return 0, fmt.Errorf("could not find domain '%s'", domain)
}

func createSDKClient(accessKeyId, accessKeySecret string) (*internal.VodClient, error) {
	clientCredentials := jdcore.NewCredentials(accessKeyId, accessKeySecret)
	client := internal.NewVodClient(clientCredentials)
	return client, nil
}
