package ksyuncdn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	ksyuncdnsdk "github.com/certimate-go/certimate/pkg/sdk3rd/ksyun/cdn"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// 金山云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 金山云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 金山云项目 ID。
	ProjectId int64 `json:"projectId,omitempty"`
	// 部署目标。
	DeployTarget string `json:"deployTarget"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 加速域名（支持泛域名）。
	Domain string `json:"domain"`
	// 证书 ID。
	// 部署目标为 [DEPLOY_TARGET_CERTIFICATE] 时必填。
	CertificateId string `json:"certificateId,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *ksyuncdnsdk.Client
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey)
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*DeployResult, error) {
	// 根据部署目标决定业务流程
	switch d.config.DeployTarget {
	case DEPLOY_TARGET_DOMAIN:
		if err := d.deployToDomain(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	case DEPLOY_TARGET_CERTIFICATE:
		if err := d.deployToCertificate(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported deploy target '%s'", d.config.DeployTarget)
	}

	return &DeployResult{}, nil
}

func (d *Deployer) deployToDomain(ctx context.Context, certPEM, privkeyPEM string) error {
	_, err := d.getAllDomains(ctx)
	if err != nil {
		return err
	}

	// 获取待部署的域名列表
	var domainIds []string
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if d.config.Domain == "" {
				return fmt.Errorf("config `domain` is required")
			}

			domainCandidates, err := d.getAllDomains(ctx)
			if err != nil {
				return err
			}
			domains := lo.Filter(domainCandidates, func(domainItem *ksyuncdnsdk.CDNDomain, _ int) bool {
				return d.config.Domain == domainItem.DomainName
			})
			if len(domains) == 0 {
				return fmt.Errorf("could not find domain")
			}

			domainIds = lo.Map(domains, func(domainItem *ksyuncdnsdk.CDNDomain, _ int) string {
				return domainItem.DomainId
			})
		}

	case DOMAIN_MATCH_PATTERN_WILDCARD:
		{
			if d.config.Domain == "" {
				return fmt.Errorf("config `domain` is required")
			}

			domainCandidates, err := d.getAllDomains(ctx)
			if err != nil {
				return err
			}

			domains := lo.Filter(domainCandidates, func(domainItem *ksyuncdnsdk.CDNDomain, _ int) bool {
				return xcerthostname.IsMatch(d.config.Domain, domainItem.DomainName)
			})
			if len(domains) == 0 {
				return fmt.Errorf("could not find any domains matched by wildcard")
			}

			domainIds = lo.Map(domains, func(domainItem *ksyuncdnsdk.CDNDomain, _ int) string {
				return domainItem.DomainId
			})
		}

	case DOMAIN_MATCH_PATTERN_CERTSAN:
		{
			domainCandidates, err := d.getAllDomains(ctx)
			if err != nil {
				return err
			}

			domains := lo.Filter(domainCandidates, func(domainItem *ksyuncdnsdk.CDNDomain, _ int) bool {
				return xcerthostname.IsMatchByCertificatePEM(certPEM, domainItem.DomainName)
			})
			if len(domains) == 0 {
				return fmt.Errorf("could not find any domains matched by certificate")
			}

			domainIds = lo.Map(domains, func(domainItem *ksyuncdnsdk.CDNDomain, _ int) string {
				return domainItem.DomainId
			})
		}

	default:
		return fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 遍历更新域名证书
	if len(domainIds) == 0 {
		d.logger.Info("no cdn domains to deploy")
	} else {
		d.logger.Info("found cdn domains to deploy", slog.Any("domainIds", domainIds))
		var errs []error

		for _, domainId := range domainIds {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := d.updateDomainCertificate(ctx, domainId, certPEM, privkeyPEM); err != nil {
					errs = append(errs, err)
				}
			}
		}

		if len(errs) > 0 {
			return errors.Join(errs...)
		}
	}

	return nil
}

func (d *Deployer) deployToCertificate(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.CertificateId == "" {
		return fmt.Errorf("config `certificateId` is required")
	}

	// 更新证书
	// REF: https://docs.ksyun.com/documents/259
	setCertificateReq := &ksyuncdnsdk.SetCertificateRequest{
		CertificateId:     lo.ToPtr(d.config.CertificateId),
		CertificateName:   lo.ToPtr(fmt.Sprintf("certimate-%d", time.Now().UnixMilli())),
		ServerCertificate: lo.ToPtr(certPEM),
		PrivateKey:        lo.ToPtr(privkeyPEM),
	}
	setCertificateResp, err := d.sdkClient.SetCertificateWithContext(ctx, setCertificateReq)
	d.logger.Debug("sdk request 'cdn.SetCertificate'", slog.Any("request", setCertificateReq), slog.Any("response", setCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'cdn.SetCertificate': %w", err)
	}

	return nil
}

func (d *Deployer) getAllDomains(ctx context.Context) ([]*ksyuncdnsdk.CDNDomain, error) {
	domains := make([]*ksyuncdnsdk.CDNDomain, 0)

	// 查询域名列表
	// REF: https://docs.ksyun.com/documents/198
	getCdnDomainsPageNumber := 1
	getCdnDomainsPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		getCdnDomainsReq := &ksyuncdnsdk.GetCDNDomainsRequest{
			ProjectId:  lo.IfF(d.config.ProjectId != 0, func() *int64 { return lo.ToPtr(d.config.ProjectId) }).Else(nil),
			PageNumber: lo.ToPtr(int32(getCdnDomainsPageNumber)),
			PageSize:   lo.ToPtr(int32(getCdnDomainsPageSize)),
		}
		getCdnDomainsResp, err := d.sdkClient.GetCDNDomainsWithContext(ctx, getCdnDomainsReq)
		d.logger.Debug("sdk request 'cdn.GetCdnDomains'", slog.Any("request", getCdnDomainsReq), slog.Any("response", getCdnDomainsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'cdn.GetCdnDomains': %w", err)
		}

		if getCdnDomainsResp.Domains == nil {
			break
		}

		ignoredStatuses := []string{"offline", "icp_checking", "icp_check_failed", "locking", "locked"}
		for _, domainItem := range getCdnDomainsResp.Domains {
			if lo.Contains(ignoredStatuses, domainItem.DomainStatus) {
				continue
			}

			domains = append(domains, domainItem)
		}

		if len(getCdnDomainsResp.Domains) < getCdnDomainsPageSize {
			break
		}

		getCdnDomainsPageNumber++
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, cloudDomainId string, certPEM, privkeyPEM string) error {
	// 为加速域名配置证书接口
	// REF: https://docs.ksyun.com/documents/261
	configCertificateReq := &ksyuncdnsdk.ConfigCertificateRequest{
		Enable:            lo.ToPtr("on"),
		DomainIds:         lo.ToPtr(cloudDomainId),
		CertificateName:   lo.ToPtr(fmt.Sprintf("certimate-%d", time.Now().UnixMilli())),
		ServerCertificate: lo.ToPtr(certPEM),
		PrivateKey:        lo.ToPtr(privkeyPEM),
	}
	configCertificateResp, err := d.sdkClient.ConfigCertificateWithContext(ctx, configCertificateReq)
	d.logger.Debug("sdk request 'cdn.ConfigCertificate'", slog.Any("request", configCertificateReq), slog.Any("response", configCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'cdn.ConfigCertificate': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, secretAccessKey string) (*ksyuncdnsdk.Client, error) {
	client, err := ksyuncdnsdk.NewClient(
		ksyuncdnsdk.WithAkSk(accessKeyId, secretAccessKey),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
