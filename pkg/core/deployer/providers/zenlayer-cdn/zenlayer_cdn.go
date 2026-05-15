package zenlayercdn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/lo"
	zcommon "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/zenlayer-cdn"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	zcdnsdk "github.com/certimate-go/certimate/pkg/sdk3rd/zenlayer/cdn"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type DeployerConfig struct {
	// Zenlayer AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// Zenlayer AccessKeyPassword。
	AccessKeyPassword string `json:"accessKeyPassword"`
	// Zenlayer 资源组 ID。
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// 全球网络加速器 ID。
	// 部署资源类型为 [RESOURCE_TYPE_ACCELERATOR] 时必填。
	AcceleratorId string `json:"acceleratorId,omitempty"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 加速域名（支持泛域名）。
	Domain string `json:"domain"`
	// 证书 ID。
	// 部署资源类型为 [RESOURCE_TYPE_CERTIFICATE] 时必填。
	CertificateId string `json:"certificateId,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *zcdnsdk.Client
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeyPassword)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		AccessKeyId:       config.AccessKeyId,
		AccessKeyPassword: config.AccessKeyPassword,
		ResourceGroupId:   config.ResourceGroupId,
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
	// 根据部署资源类型决定部署方式
	switch d.config.ResourceType {
	case RESOURCE_TYPE_DOMAIN:
		if err := d.deployToDomain(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	case RESOURCE_TYPE_CERTIFICATE:
		if err := d.deployToCertificate(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported resource type '%s'", d.config.ResourceType)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) deployToDomain(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.Domain == "" {
		return errors.New("config `domain` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 获取待部署的域名列表
	domainIds := make([]string, 0)
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if d.config.Domain == "" {
				return errors.New("config `domain` is required")
			}

			domainCandidates, err := d.getAllDomains(ctx)
			if err != nil {
				return err
			}
			domains := lo.Filter(domainCandidates, func(domainItem *zcdnsdk.DomainInfo, _ int) bool {
				return domainItem.DomainName == d.config.Domain
			})
			if len(domains) == 0 {
				return errors.New("could not find domain")
			}

			domainIds = lo.Map(domains, func(domainItem *zcdnsdk.DomainInfo, _ int) string {
				return domainItem.DomainId
			})
		}

	case DOMAIN_MATCH_PATTERN_WILDCARD:
		{
			if d.config.Domain == "" {
				return errors.New("config `domain` is required")
			}

			domainCandidates, err := d.getAllDomains(ctx)
			if err != nil {
				return err
			}

			domains := lo.Filter(domainCandidates, func(domainItem *zcdnsdk.DomainInfo, _ int) bool {
				return xcerthostname.IsMatch(d.config.Domain, domainItem.DomainName)
			})
			if len(domains) == 0 {
				return errors.New("could not find any domains matched by wildcard")
			}

			domainIds = lo.Map(domains, func(domainItem *zcdnsdk.DomainInfo, _ int) string {
				return domainItem.DomainId
			})
		}

	case DOMAIN_MATCH_PATTERN_CERTSAN:
		{
			domainCandidates, err := d.getAllDomains(ctx)
			if err != nil {
				return err
			}

			domains := lo.Filter(domainCandidates, func(domainItem *zcdnsdk.DomainInfo, _ int) bool {
				return xcerthostname.IsMatchByCertificatePEM(certPEM, domainItem.DomainName)
			})
			if len(domains) == 0 {
				return errors.New("could not find any domains matched by certificate")
			}

			domainIds = lo.Map(domains, func(domainItem *zcdnsdk.DomainInfo, _ int) string {
				return domainItem.DomainId
			})
		}

	default:
		return fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 遍历绑定证书
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
				if err := d.updateDomainCertificate(ctx, domainId, upres.CertId); err != nil {
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
		return errors.New("config `certificateId` is required")
	}

	// 替换证书
	opres, err := d.sdkCertmgr.Replace(ctx, d.config.CertificateId, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to replace certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate replaced", slog.Any("result", opres))
	}

	return nil
}

func (d *Deployer) getAllDomains(ctx context.Context) ([]*zcdnsdk.DomainInfo, error) {
	domains := make([]*zcdnsdk.DomainInfo, 0)

	// 查询加速域名列表
	// https://docs.console.zenlayer.com/api-reference/cn/networking/cdn/domain/describedomains
	describeDomainsPageNum := 1
	describeDomainsPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		describeDomainsReq := zcdnsdk.NewDescribeDomainsRequest()
		describeDomainsReq.DomainStatus = "ENABLED"
		describeDomainsReq.PageNum = describeDomainsPageNum
		describeDomainsReq.PageSize = describeDomainsPageSize
		describeDomainsResp, err := d.sdkClient.DescribeDomains(describeDomainsReq)
		d.logger.Debug("sdk request 'cdn.DescribeDomains'", slog.Any("request", describeDomainsReq), slog.Any("response", describeDomainsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'cdn.DescribeDomains': %w", err)
		}

		for _, domainItem := range describeDomainsResp.Response.DataSet {
			domains = append(domains, domainItem)
		}

		if len(describeDomainsResp.Response.DataSet) < describeDomainsPageSize {
			break
		}

		describeDomainsPageNum++
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, cloudDomainId string, cloudCertId string) error {
	// 查询加速域名的 SSL 证书
	// REF: https://docs.console.zenlayer.com/api-reference/cn/networking/cdn/domain/describedomaincertificate
	describeDomainCertificateReq := zcdnsdk.NewDescribeDomainCertificateRequest()
	describeDomainCertificateReq.DomainId = cloudDomainId
	describeDomainCertificateResp, err := d.sdkClient.DescribeDomainCertificate(describeDomainCertificateReq)
	d.logger.Debug("sdk request 'cdn.DescribeDomainCertificate'", slog.Any("request", describeDomainCertificateReq), slog.Any("response", describeDomainCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'cdn.DescribeDomainCertificate': %w", err)
	} else if describeDomainCertificateResp.Response.Certificate != nil && describeDomainCertificateResp.Response.Certificate.CertificateId == cloudCertId {
		return nil
	}

	// 修改加速域名的 SSL 证书
	// REF: https://docs.console.zenlayer.com/api-reference/cn/networking/cdn/domain/modifydomaincertificate
	modifyDomainCertificateReq := zcdnsdk.NewModifyDomainCertificateRequest()
	modifyDomainCertificateReq.DomainId = cloudDomainId
	modifyDomainCertificateReq.CertificateId = cloudCertId
	modifyDomainCertificateResp, err := d.sdkClient.ModifyDomainCertificate(modifyDomainCertificateReq)
	d.logger.Debug("sdk request 'cdn.ModifyDomainCertificate'", slog.Any("request", modifyDomainCertificateReq), slog.Any("response", modifyDomainCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'cdn.ModifyDomainCertificate': %w", err)
	}

	// 查询加速域名状态，等待部署状态变更
	// REF: https://docs.console.zenlayer.com/api-reference/cn/networking/cdn/domain/describedomains
	if _, err := xwait.UntilWithContext(ctx, func(_ context.Context, _ int) (bool, error) {
		describeDomainsReq := zcdnsdk.NewDescribeDomainsRequest()
		describeDomainsReq.DomainIds = []string{cloudDomainId}
		describeDomainsReq.PageNum = 1
		describeDomainsReq.PageSize = 1
		describeDomainsResp, err := d.sdkClient.DescribeDomains(describeDomainsReq)
		d.logger.Debug("sdk request 'cdn.DescribeDomains'", slog.Any("request", describeDomainsReq), slog.Any("response", describeDomainsResp))
		if err != nil {
			return false, fmt.Errorf("failed to execute sdk request 'cdn.DescribeDomains': %w", err)
		} else if len(describeDomainsResp.Response.DataSet) == 0 {
			return false, fmt.Errorf("could not found domain '%s'", cloudDomainId)
		}

		switch describeDomainsResp.Response.DataSet[0].ConfigStatus {
		case "DEPLOYED":
			return true, nil
		case "FAILED":
			return false, fmt.Errorf("unexpected zenlayer domain status")
		}

		d.logger.Info("waiting for zenlayer domain deploying completion ...")
		return false, nil
	}, time.Second*5); err != nil {
		return err
	}

	return nil
}

func createSDKClient(accessKeyId, accessKeyPassword string) (*zcdnsdk.Client, error) {
	config := zcommon.NewConfig()

	client, err := zcdnsdk.NewClient(config, accessKeyId, accessKeyPassword)
	if err != nil {
		return nil, err
	}

	return client, nil
}
