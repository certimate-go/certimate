package volcenginecdn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	vecdn "github.com/volcengine/volcengine-go-sdk/service/cdn"
	ve "github.com/volcengine/volcengine-go-sdk/volcengine"
	vesession "github.com/volcengine/volcengine-go-sdk/volcengine/session"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/volcengine-cdn"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/volcengine-cdn/internal"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
)

type DeployerConfig struct {
	// 火山引擎 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 火山引擎 AccessKeySecret。
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
	sdkClient  *internal.CdnClient
	sdkCertmgr certmgr.Provider
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

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 获取待部署的域名列表
	domains := make([]string, 0)
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
				domainCandidates, err := d.getMatchedDomainsByWildcard(ctx, d.config.Domain)
				if err != nil {
					return nil, err
				}

				domains = domainCandidates
			} else {
				domains = []string{d.config.Domain}
			}
		}

	case DOMAIN_MATCH_PATTERN_CERTSAN:
		{
			domainCandidates, err := d.getMatchedDomainsByCertId(ctx, upres.CertId)
			if err != nil {
				return nil, err
			}

			domains = domainCandidates
		}

	default:
		return nil, fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 遍历绑定证书
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

func (d *Deployer) getMatchedDomainsByWildcard(ctx context.Context, wildcardDomain string) ([]string, error) {
	domains := make([]string, 0)

	// 查询加速域名列表，获取匹配的域名
	// REF: https://www.volcengine.com/docs/6454/75269
	listCdnDomainsPageNum := 1
	listCdnDomainsPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCdnDomainsReq := &vecdn.ListCdnDomainsInput{
			Domain:   ve.String(strings.TrimPrefix(wildcardDomain, "*.")),
			Status:   ve.String("online"),
			PageNum:  ve.Int64(int64(listCdnDomainsPageNum)),
			PageSize: ve.Int64(int64(listCdnDomainsPageSize)),
		}
		listCdnDomainsResp, err := d.sdkClient.ListCdnDomains(listCdnDomainsReq)
		d.logger.Debug("sdk request 'cdn.ListCdnDomains'", slog.Any("request", listCdnDomainsReq), slog.Any("response", listCdnDomainsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'cdn.ListCdnDomains': %w", err)
		}

		for _, domainItem := range listCdnDomainsResp.Data {
			if xcerthostname.IsMatch(wildcardDomain, ve.StringValue(domainItem.Domain)) {
				domains = append(domains, ve.StringValue(domainItem.Domain))
			}
		}

		if len(listCdnDomainsResp.Data) < listCdnDomainsPageSize {
			break
		}

		listCdnDomainsPageSize++
	}

	if len(domains) == 0 {
		return nil, errors.New("could not find any domains matched by wildcard")
	}

	return domains, nil
}

func (d *Deployer) getMatchedDomainsByCertId(ctx context.Context, cloudCertId string) ([]string, error) {
	domains := make([]string, 0)

	// 获取指定证书可关联的域名
	// REF: https://www.volcengine.com/docs/6454/125711
	describeCertConfigReq := &vecdn.DescribeCertConfigInput{
		CertId: ve.String(cloudCertId),
	}
	describeCertConfigResp, err := d.sdkClient.DescribeCertConfig(describeCertConfigReq)
	d.logger.Debug("sdk request 'cdn.DescribeCertConfig'", slog.Any("request", describeCertConfigReq), slog.Any("response", describeCertConfigResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cdn.DescribeCertConfig': %w", err)
	}

	if describeCertConfigResp.CertNotConfig != nil {
		for i := range describeCertConfigResp.CertNotConfig {
			domains = append(domains, ve.StringValue(describeCertConfigResp.CertNotConfig[i].Domain))
		}
	}

	if describeCertConfigResp.OtherCertConfig != nil {
		for i := range describeCertConfigResp.OtherCertConfig {
			domains = append(domains, ve.StringValue(describeCertConfigResp.OtherCertConfig[i].Domain))
		}
	}

	if len(domains) == 0 {
		if len(describeCertConfigResp.SpecifiedCertConfig) == 0 {
			return nil, errors.New("could not find any domains matched by certificate")
		}
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain string, cloudCertId string) error {
	// 关联证书与加速域名
	// REF: https://www.volcengine.com/docs/6454/125712
	batchDeployCertReq := &vecdn.BatchDeployCertInput{
		Domain: ve.String(domain),
		CertId: ve.String(cloudCertId),
	}
	batchDeployCertResp, err := d.sdkClient.BatchDeployCert(batchDeployCertReq)
	d.logger.Debug("sdk request 'cdn.BatchDeployCert'", slog.Any("request", batchDeployCertReq), slog.Any("response", batchDeployCertResp))
	if err != nil {
		return err
	}

	return nil
}

func createSDKClient(accessKeyId, accessKeySecret string) (*internal.CdnClient, error) {
	config := ve.NewConfig().
		WithAkSk(accessKeyId, accessKeySecret).
		WithRegion("cn-north-1")

	session, err := vesession.NewSession(config)
	if err != nil {
		return nil, err
	}

	client := internal.NewCdnClient(session)
	return client, nil
}
