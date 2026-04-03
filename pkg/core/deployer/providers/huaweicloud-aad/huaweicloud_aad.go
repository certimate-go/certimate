package huaweicloudaad

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	hcaad "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/aad/v1"
	hcaadmodelv1 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/aad/v1/model"
	hcaadregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/aad/v1/region"
	hcaadmodelv2 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/aad/v2/model"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/huaweicloud-aad/internal"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
)

type DeployerConfig struct {
	// 华为云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 华为云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 华为云企业项目 ID。
	EnterpriseProjectId string `json:"enterpriseProjectId,omitempty"`
	// DDoS 高防实例 ID。
	InstanceId string `json:"instanceId"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 网站域名（支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *internal.AadClient
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(
		config.AccessKeyId,
		config.SecretAccessKey,
	)
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	if d.config.InstanceId == "" {
		return nil, errors.New("config `instanceId` is required")
	}

	// 获取待部署的域名列表
	var domainIds []string
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if d.config.Domain == "" {
				return nil, errors.New("config `domain` is required")
			}

			domainCandidates, err := d.getAllDomainsByInstanceId(ctx, d.config.InstanceId)
			if err != nil {
				return nil, err
			}
			domains := lo.Filter(domainCandidates, func(domainItem *hcaadmodelv2.InstanceDomainItem, _ int) bool {
				return lo.FromPtr(domainItem.DomainName) == d.config.Domain
			})
			if len(domains) == 0 {
				return nil, errors.New("could not find domain")
			}

			domainIds = lo.Map(domains, func(domainItem *hcaadmodelv2.InstanceDomainItem, _ int) string {
				return lo.FromPtr(domainItem.DomainId)
			})
		}

	case DOMAIN_MATCH_PATTERN_WILDCARD:
		{
			if d.config.Domain == "" {
				return nil, errors.New("config `domain` is required")
			}

			domainCandidates, err := d.getAllDomainsByInstanceId(ctx, d.config.InstanceId)
			if err != nil {
				return nil, err
			}

			domains := lo.Filter(domainCandidates, func(domainItem *hcaadmodelv2.InstanceDomainItem, _ int) bool {
				return xcerthostname.IsMatch(d.config.Domain, lo.FromPtr(domainItem.DomainName)) ||
					strings.TrimPrefix(d.config.Domain, "*") == strings.TrimPrefix(lo.FromPtr(domainItem.DomainName), "*")
			})
			if len(domains) == 0 {
				return nil, errors.New("could not find any domains matched by wildcard")
			}

			domainIds = lo.Map(domains, func(domainItem *hcaadmodelv2.InstanceDomainItem, _ int) string {
				return lo.FromPtr(domainItem.DomainId)
			})
		}

	case DOMAIN_MATCH_PATTERN_CERTSAN:
		{
			certX509, err := xcert.ParseCertificateFromPEM(certPEM)
			if err != nil {
				return nil, err
			}

			domainCandidates, err := d.getAllDomainsByInstanceId(ctx, d.config.InstanceId)
			if err != nil {
				return nil, err
			}

			domains := lo.Filter(domainCandidates, func(domainItem *hcaadmodelv2.InstanceDomainItem, _ int) bool {
				return certX509.VerifyHostname(lo.FromPtr(domainItem.DomainName)) == nil
			})
			if len(domains) == 0 {
				return nil, errors.New("could not find any domains matched by certificate")
			}

			domainIds = lo.Map(domains, func(domainItem *hcaadmodelv2.InstanceDomainItem, _ int) string {
				return lo.FromPtr(domainItem.DomainId)
			})
		}

	default:
		return nil, fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 遍历更新域名证书
	if len(domainIds) == 0 {
		d.logger.Info("no aad domains to deploy")
	} else {
		d.logger.Info("found aad domains to deploy", slog.Any("domainIds", domainIds))
		var errs []error

		for _, domainId := range domainIds {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				if err := d.updateDomainCertificate(ctx, domainId, certPEM, privkeyPEM); err != nil {
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

func (d *Deployer) getAllDomainsByInstanceId(ctx context.Context, cloudInstanceId string) ([]*hcaadmodelv2.InstanceDomainItem, error) {
	domains := make([]*hcaadmodelv2.InstanceDomainItem, 0)

	// 查询实例关联的域名信息
	// REF: https://support.huaweicloud.com/intl/zh-cn/api-aad/ListInstanceDomains.html
	listInstanceDomainsOffset := 0
	listInstanceDomainsLimit := 10
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listInstanceDomainsReq := &hcaadmodelv2.ListInstanceDomainsRequest{
			InstanceId: cloudInstanceId,
			Offset:     lo.ToPtr(int32(listInstanceDomainsOffset)),
			Limit:      lo.ToPtr(int32(listInstanceDomainsLimit)),
		}
		listInstanceDomainsResp, err := d.sdkClient.ListInstanceDomains(listInstanceDomainsReq)
		d.logger.Debug("sdk request 'aad.ListInstanceDomains'", slog.Any("request", listInstanceDomainsReq), slog.Any("response", listInstanceDomainsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'aad.ListInstanceDomains': %w", err)
		}

		if listInstanceDomainsResp.Domains == nil {
			break
		}

		ignoredStatuses := []string{"1"}
		for _, domainItem := range *listInstanceDomainsResp.Domains {
			if lo.Contains(ignoredStatuses, lo.FromPtr(domainItem.DomainStatus)) {
				continue
			}

			domains = append(domains, &domainItem)
		}

		if len(*listInstanceDomainsResp.Domains) < listInstanceDomainsLimit {
			break
		}

		listInstanceDomainsOffset += listInstanceDomainsLimit
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domainId string, certPEM, privkeyPEM string) error {
	// 上传域名对应证书
	// REF: https://support.huaweicloud.com/intl/zh-cn/api-aad/SetCertForDomain.html
	setCertForDomainReq := &hcaadmodelv1.SetCertForDomainRequest{
		Body: &hcaadmodelv1.CertificateBody{
			OpType:      0,
			DomainId:    domainId,
			CertName:    fmt.Sprintf("certimate_%d", time.Now().UnixMilli()),
			CertFile:    lo.ToPtr(certPEM),
			CertKeyFile: lo.ToPtr(privkeyPEM),
		},
	}
	setCertForDomainResp, err := d.sdkClient.SetCertForDomain(setCertForDomainReq)
	d.logger.Debug("sdk request 'aad.SetCertForDomain'", slog.Any("request", setCertForDomainReq), slog.Any("response", setCertForDomainResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'aad.SetCertForDomain': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, secretAccessKey string) (*internal.AadClient, error) {
	region := "cn-north-4" // AAD 服务默认区域：华北北京四

	auth, err := global.NewCredentialsBuilder().
		WithAk(accessKeyId).
		WithSk(secretAccessKey).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	hcRegion, err := hcaadregion.SafeValueOf(region)
	if err != nil {
		return nil, err
	}

	hcClient, err := hcaad.AadClientBuilder().
		WithRegion(hcRegion).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	client := internal.NewAadClient(hcClient)
	return client, nil
}
