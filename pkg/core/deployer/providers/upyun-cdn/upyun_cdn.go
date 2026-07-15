package upyuncdn

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/upyun-ssl"
	upyunsdk "github.com/certimate-go/certimate/pkg/sdk3rd/upyun/console"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// 又拍云账号用户名。
	Username string `json:"username"`
	// 又拍云账号密码。
	Password string `json:"password"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 加速域名（支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *upyunsdk.Client
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.Username, config.Password)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		Username: config.Username,
		Password: config.Password,
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
			return d.updateDomainCertificate(ctx, domain, upres.CertId)
		}); err != nil {
			return nil, err
		}
	}

	return &DeployResult{}, nil
}

func (d *Deployer) getAllDomains(ctx context.Context) ([]string, error) {
	domains := make([]string, 0)

	// 获取服务列表
	getBucketsPage := 1
	getBucketsPerPage := 10
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		getBucketsReq := &upyunsdk.GetBucketsRequest{
			Type:          "ucdn",
			Tag:           "all",
			Status:        "all",
			IsSecurityCDN: false,
			WithDomains:   true,
			Page:          int32(getBucketsPage),
			PerPage:       int32(getBucketsPerPage),
		}
		getBucketsResp, err := d.sdkClient.GetBucketsWithContext(ctx, getBucketsReq)
		d.logger.Debug("sdk request 'console.GetBuckets'", slog.Any("request", getBucketsReq), slog.Any("response", getBucketsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'console.GetBuckets': %w", err)
		}

		if getBucketsResp.Data == nil {
			break
		}

		for _, bucketItem := range getBucketsResp.Data.Buckets {
			if !bucketItem.Visible {
				continue
			}

			for _, domainItem := range bucketItem.Domains {
				if strings.EqualFold(domainItem.Status, "NORMAL") && !strings.HasSuffix(domainItem.Domain, ".test.upcdn.net") {
					domains = append(domains, domainItem.Domain)
				}
			}
		}

		if len(getBucketsResp.Data.Buckets) < getBucketsPerPage {
			break
		}

		getBucketsPage++
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain string, cloudCertId string) error {
	// 获取域名证书配置
	getHttpsServiceManagerResp, err := d.sdkClient.GetHttpsServiceManagerWithContext(ctx, domain)
	d.logger.Debug("sdk request 'console.GetHttpsServiceManager'", slog.String("params.domain", domain), slog.Any("response", getHttpsServiceManagerResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'console.GetHttpsServiceManager': %w", err)
	}

	// 判断域名是否已启用 HTTPS
	// 如果已启用，迁移域名证书；否则，设置新证书
	_, lastCertIndex, _ := lo.FindIndexOf(getHttpsServiceManagerResp.Data.Domains, func(item upyunsdk.HttpsServiceManagerDomain) bool {
		return item.Https
	})
	if lastCertIndex == -1 {
		updateHttpsCertificateManagerReq := &upyunsdk.UpdateHttpsCertificateManagerRequest{
			CertificateId: cloudCertId,
			Domain:        domain,
			Https:         true,
			ForceHttps:    true,
		}
		updateHttpsCertificateManagerResp, err := d.sdkClient.UpdateHttpsCertificateManagerWithContext(ctx, updateHttpsCertificateManagerReq)
		d.logger.Debug("sdk request 'console.EnableDomainHttps'", slog.Any("request", updateHttpsCertificateManagerReq), slog.Any("response", updateHttpsCertificateManagerResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'console.UpdateHttpsCertificateManager': %w", err)
		}
	} else if getHttpsServiceManagerResp.Data.Domains[lastCertIndex].CertificateId != cloudCertId {
		migrateHttpsDomainReq := &upyunsdk.MigrateHttpsDomainRequest{
			CertificateId: cloudCertId,
			Domain:        domain,
		}
		migrateHttpsDomainResp, err := d.sdkClient.MigrateHttpsDomainWithContext(ctx, migrateHttpsDomainReq)
		d.logger.Debug("sdk request 'console.MigrateHttpsDomain'", slog.Any("request", migrateHttpsDomainReq), slog.Any("response", migrateHttpsDomainResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'console.MigrateHttpsDomain': %w", err)
		}
	}

	return nil
}

func createSDKClient(username, password string) (*upyunsdk.Client, error) {
	client, err := upyunsdk.NewClient(
		upyunsdk.WithLogins(username, password),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
