package byteplusapig

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	bp "github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus"
	bpsession "github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/session"
	"github.com/samber/lo"

	bpapig "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/byteplus-sdk/byteplus-go-sdk-v2/service/apig"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/byteplus-certcenter"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// BytePlus AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// BytePlus SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// BytePlus 项目名称。
	ProjectName string `json:"projectName,omitempty"`
	// BytePlus 地域。
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
	sdkClient  *bpapig.APIG
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		SecretAccessKey: config.SecretAccessKey,
		ProjectName:     config.ProjectName,
		Region:          "ap-singapore-1",
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
	domainIds := make([]string, 0)
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if d.config.Domain == "" {
				return nil, fmt.Errorf("config `domain` is required")
			}

			domainCandidates, err := d.getAllDomains(ctx)
			if err != nil {
				return nil, err
			}

			domains := lo.Filter(domainCandidates, func(domainItem *bpapig.ItemForListCustomDomainsOutput, _ int) bool {
				return d.config.Domain == lo.FromPtr(domainItem.Domain)
			})
			if len(domains) == 0 {
				return nil, fmt.Errorf("could not find domain")
			}

			domainIds = lo.Map(domains, func(domainItem *bpapig.ItemForListCustomDomainsOutput, _ int) string {
				return lo.FromPtr(domainItem.Id)
			})
		}

	case DOMAIN_MATCH_PATTERN_WILDCARD:
		{
			if d.config.Domain == "" {
				return nil, fmt.Errorf("config `domain` is required")
			}

			domainCandidates, err := d.getAllDomains(ctx)
			if err != nil {
				return nil, err
			}

			domains := lo.Filter(domainCandidates, func(domainItem *bpapig.ItemForListCustomDomainsOutput, _ int) bool {
				return xcerthostname.IsMatch(d.config.Domain, lo.FromPtr(domainItem.Domain))
			})
			if len(domains) == 0 {
				return nil, fmt.Errorf("could not find any domains matched by wildcard")
			}

			domainIds = lo.Map(domains, func(domainItem *bpapig.ItemForListCustomDomainsOutput, _ int) string {
				return lo.FromPtr(domainItem.Id)
			})
		}

	case DOMAIN_MATCH_PATTERN_CERTSAN:
		{
			domainCandidates, err := d.getAllDomains(ctx)
			if err != nil {
				return nil, err
			}

			domains := lo.Filter(domainCandidates, func(domainItem *bpapig.ItemForListCustomDomainsOutput, _ int) bool {
				return xcerthostname.IsMatchByCertificatePEM(certPEM, lo.FromPtr(domainItem.Domain))
			})
			if len(domains) == 0 {
				return nil, fmt.Errorf("could not find any domains matched by certificate")
			}

			domainIds = lo.Map(domains, func(domainItem *bpapig.ItemForListCustomDomainsOutput, _ int) string {
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

	return &DeployResult{}, nil
}

func (d *Deployer) getAllDomains(ctx context.Context) ([]*bpapig.ItemForListCustomDomainsOutput, error) {
	domains := make([]*bpapig.ItemForListCustomDomainsOutput, 0)

	// 查询自定义域名列表
	listCustomDomainsPageNumber := 1
	listCustomDomainsPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCustomDomainsReq := &bpapig.ListCustomDomainsInput{
			PageNumber: bp.Int64(int64(listCustomDomainsPageNumber)),
			PageSize:   bp.Int64(int64(listCustomDomainsPageSize)),
		}
		listCustomDomainsResp, err := d.sdkClient.ListCustomDomainsWithContext(ctx, listCustomDomainsReq)
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
	getCustomDomainReq := &bpapig.GetCustomDomainInput{
		Id: bp.String(cloudDomainId),
	}
	getCustomDomainResp, err := d.sdkClient.GetCustomDomainWithContext(ctx, getCustomDomainReq)
	d.logger.Debug("sdk request 'apig.GetCustomDomain'", slog.Any("request", getCustomDomainReq), slog.Any("response", getCustomDomainResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'apig.GetCustomDomain': %w", err)
	}

	// 更新自定义域名
	updateCustomDomainReq := &bpapig.UpdateCustomDomainInput{
		Id:            bp.String(cloudDomainId),
		Protocol:      getCustomDomainResp.CustomDomain.Protocol,
		CertificateId: bp.String(cloudCertId),
	}
	if !lo.Contains(bp.StringValueSlice(updateCustomDomainReq.Protocol), "HTTPS") {
		updateCustomDomainReq.Protocol = append(updateCustomDomainReq.Protocol, bp.String("HTTPS"))
	}
	updateCustomDomainResp, err := d.sdkClient.UpdateCustomDomainWithContext(ctx, updateCustomDomainReq)
	d.logger.Debug("sdk request 'apig.UpdateCustomDomain'", slog.Any("request", updateCustomDomainReq), slog.Any("response", updateCustomDomainResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'apig.UpdateCustomDomain': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*bpapig.APIG, error) {
	config := bp.NewConfig().
		WithAkSk(accessKeyId, secretAccessKey).
		WithRegion(region)

	session, err := bpsession.NewSession(config)
	if err != nil {
		return nil, err
	}

	client := bpapig.New(session)
	return client, nil
}
