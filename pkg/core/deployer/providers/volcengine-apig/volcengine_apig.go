package volcengineapig

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/samber/lo"
	veapig "github.com/volcengine/volcengine-go-sdk/service/apig"
	ve "github.com/volcengine/volcengine-go-sdk/volcengine"
	vesession "github.com/volcengine/volcengine-go-sdk/volcengine/session"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/volcengine-certcenter"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/volcengine-apig/internal"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
)

type DeployerConfig struct {
	// 火山引擎 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 火山引擎 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 火山引擎地域。
	Region string `json:"region"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 加速域名（支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *internal.ApigClient
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
		Region:          config.Region,
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
	domainIds := make([]string, 0)
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if d.config.Domain == "" {
				return nil, errors.New("config `domain` is required")
			}

			domainCandidates, err := d.getAllDomains(ctx)
			if err != nil {
				return nil, err
			}
			domains := lo.Filter(domainCandidates, func(domainItem *veapig.ItemForListCustomDomainsOutput, _ int) bool {
				return lo.FromPtr(domainItem.Domain) == d.config.Domain
			})
			if len(domains) == 0 {
				return nil, errors.New("could not find domain")
			}

			domainIds = lo.Map(domains, func(domainItem *veapig.ItemForListCustomDomainsOutput, _ int) string {
				return lo.FromPtr(domainItem.Id)
			})
		}

	case DOMAIN_MATCH_PATTERN_WILDCARD:
		{
			if d.config.Domain == "" {
				return nil, errors.New("config `domain` is required")
			}

			domainCandidates, err := d.getAllDomains(ctx)
			if err != nil {
				return nil, err
			}

			domains := lo.Filter(domainCandidates, func(domainItem *veapig.ItemForListCustomDomainsOutput, _ int) bool {
				return xcerthostname.IsMatch(d.config.Domain, lo.FromPtr(domainItem.Domain))
			})
			if len(domains) == 0 {
				return nil, errors.New("could not find any domains matched by wildcard")
			}

			domainIds = lo.Map(domains, func(domainItem *veapig.ItemForListCustomDomainsOutput, _ int) string {
				return lo.FromPtr(domainItem.Id)
			})
		}

	case DOMAIN_MATCH_PATTERN_CERTSAN:
		{
			domainCandidates, err := d.getAllDomains(ctx)
			if err != nil {
				return nil, err
			}

			domains := lo.Filter(domainCandidates, func(domainItem *veapig.ItemForListCustomDomainsOutput, _ int) bool {
				return xcerthostname.IsMatchByCertificatePEM(certPEM, lo.FromPtr(domainItem.Domain))
			})
			if len(domains) == 0 {
				return nil, errors.New("could not find any domains matched by certificate")
			}

			domainIds = lo.Map(domains, func(domainItem *veapig.ItemForListCustomDomainsOutput, _ int) string {
				return lo.FromPtr(domainItem.Id)
			})
		}

	default:
		return nil, fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 遍历更新域名证书
	if len(domainIds) == 0 {
		d.logger.Info("no apig domains to deploy")
	} else {
		d.logger.Info("found apig domains to deploy", slog.Any("domainIds", domainIds))
		var errs []error

		for _, domainId := range domainIds {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				if err := d.updateDomainCertificate(ctx, domainId, upres.CertId); err != nil {
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

func (d *Deployer) getAllDomains(ctx context.Context) ([]*veapig.ItemForListCustomDomainsOutput, error) {
	domains := make([]*veapig.ItemForListCustomDomainsOutput, 0)

	// 查询自定义域名列表
	// https://api.volcengine.com/api-explorer?action=ListCustomDomains&serviceCode=apig&version=2021-03-03
	listCustomDomainsPageNumber := 1
	listCustomDomainsPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCustomDomainsReq := &veapig.ListCustomDomainsInput{
			PageNumber: ve.Int64(int64(listCustomDomainsPageNumber)),
			PageSize:   ve.Int64(int64(listCustomDomainsPageSize)),
		}
		listCustomDomainsResp, err := d.sdkClient.ListCustomDomains(listCustomDomainsReq)
		d.logger.Debug("sdk request 'apig.ListCustomDomains'", slog.Any("request", listCustomDomainsReq), slog.Any("response", listCustomDomainsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'apig.ListCustomDomains': %w", err)
		}

		ignoredStatuses := []string{"Creating", "CreationFailed", "Deleting", "DeletionFailed"}
		for _, domainItem := range listCustomDomainsResp.Items {
			if lo.Contains(ignoredStatuses, *domainItem.Status) {
				continue
			}

			domains = append(domains, domainItem)
		}

		if len(listCustomDomainsResp.Items) < listCustomDomainsPageSize {
			break
		}

		listCustomDomainsPageNumber++
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, cloudDomainId string, cloudCertId string) error {
	// 查询自定义域名
	// REF: https://api.volcengine.com/api-explorer?action=GetCustomDomain&serviceCode=apig&version=2021-03-03
	getCustomDomainReq := &veapig.GetCustomDomainInput{
		Id: ve.String(cloudDomainId),
	}
	getCustomDomainResp, err := d.sdkClient.GetCustomDomain(getCustomDomainReq)
	d.logger.Debug("sdk request 'apig.GetCustomDomain'", slog.Any("request", getCustomDomainReq), slog.Any("response", getCustomDomainResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'apig.GetCustomDomain': %w", err)
	}

	// 更新自定义域名
	// REF: https://api.volcengine.com/api-explorer?action=UpdateCustomDomain&serviceCode=apig&version=2021-03-03
	updateCustomDomainReq := &veapig.UpdateCustomDomainInput{
		Id:            ve.String(cloudDomainId),
		Protocol:      getCustomDomainResp.CustomDomain.Protocol,
		CertificateId: ve.String(cloudCertId),
	}
	if !lo.Contains(ve.StringValueSlice(updateCustomDomainReq.Protocol), "HTTPS") {
		updateCustomDomainReq.Protocol = append(updateCustomDomainReq.Protocol, ve.String("HTTPS"))
	}
	updateCustomDomainResp, err := d.sdkClient.UpdateCustomDomain(updateCustomDomainReq)
	d.logger.Debug("sdk request 'apig.UpdateCustomDomain'", slog.Any("request", updateCustomDomainReq), slog.Any("response", updateCustomDomainResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'apig.UpdateCustomDomain': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, accessKeySecret, region string) (*internal.ApigClient, error) {
	config := ve.NewConfig().
		WithAkSk(accessKeyId, accessKeySecret).
		WithRegion(region)

	session, err := vesession.NewSession(config)
	if err != nil {
		return nil, err
	}

	client := internal.NewApigClient(session)
	return client, nil
}
