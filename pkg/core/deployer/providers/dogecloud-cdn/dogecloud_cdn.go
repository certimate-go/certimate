package dogecloudcdn

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/dogecloud"
	dogecloudsdk "github.com/certimate-go/certimate/pkg/sdk3rd/dogecloud"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// 多吉云 AccessKey。
	AccessKey string `json:"accessKey"`
	// 多吉云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 加速域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *dogecloudsdk.Client
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKey, config.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
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

	// 获取域名列表
	// REF: https://docs.dogecloud.com/cdn/api-domain-list
	listCdnDomainResp, err := d.sdkClient.ListCdnDomainWithContext(ctx)
	d.logger.Debug("sdk request 'cdn.ListCdnDomain'", slog.Any("response", listCdnDomainResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cdn.ListCdnDomain': %w", err)
	}

	if listCdnDomainResp.Data != nil {
		ignoredStatuses := []string{"offline"}
		for _, domainItem := range listCdnDomainResp.Data.Domains {
			if lo.Contains(ignoredStatuses, domainItem.Status) {
				continue
			}

			domains = append(domains, domainItem.Name)
		}
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain string, cloudCertId int64) error {
	// 绑定证书
	// REF: https://docs.dogecloud.com/cdn/api-cert-bind
	bindCdnCertReq := &dogecloudsdk.BindCdnCertRequest{
		CertId: cloudCertId,
		Domain: domain,
	}
	bindCdnCertResp, err := d.sdkClient.BindCdnCertWithContext(ctx, bindCdnCertReq)
	d.logger.Debug("sdk request 'cdn.BindCdnCert'", slog.Any("request", bindCdnCertReq), slog.Any("response", bindCdnCertResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'cdn.BindCdnCert': %w", err)
	}

	return nil
}

func createSDKClient(accessKey, secretKey string) (*dogecloudsdk.Client, error) {
	client, err := dogecloudsdk.NewClient(
		dogecloudsdk.WithAkSk(accessKey, secretKey),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
