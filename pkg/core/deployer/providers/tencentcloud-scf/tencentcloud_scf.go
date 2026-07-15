package tencentcloudscf

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	tcscf "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/scf/v20180416"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/tencentcloud-ssl"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
	xtencentcloud "github.com/certimate-go/certimate/pkg/utils/third-party/tencentcloud"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// 腾讯云 SecretId。
	SecretId string `json:"secretId"`
	// 腾讯云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 腾讯云项目 ID。
	ProjectId int64 `json:"projectId,omitempty"`
	// 腾讯云接口端点。
	Endpoint string `json:"endpoint,omitempty"`
	// 腾讯云地域。
	Region string `json:"region"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 自定义域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *tcscf.Client
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.SecretId, config.SecretKey, config.Endpoint, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		SecretId:  config.SecretId,
		SecretKey: config.SecretKey,
		ProjectId: config.ProjectId,
		Endpoint:  lo.Ternary(xtencentcloud.IsIntlAPIEndpoint(config.Endpoint), "ssl.intl.tencentcloudapi.com", ""),
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
		d.logger.Info("no scf domains to deploy")
	} else {
		d.logger.Info("found scf domains to deploy", slog.Any("domains", domains))

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

	// 获取云函数自定义域名列表
	// REF: https://cloud.tencent.com/document/api/583/111923
	listCustomDomainsOffset := 0
	listCustomDomainsLimit := 20
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		describeLiveDomainsReq := tcscf.NewListCustomDomainsRequest()
		describeLiveDomainsReq.Offset = common.Uint64Ptr(uint64(listCustomDomainsOffset))
		describeLiveDomainsReq.Limit = common.Uint64Ptr(uint64(listCustomDomainsLimit))
		describeLiveDomainsResp, err := d.sdkClient.ListCustomDomainsWithContext(ctx, describeLiveDomainsReq)
		d.logger.Debug("sdk request 'scf.DescribeLiveDomains'", slog.Any("request", describeLiveDomainsReq), slog.Any("response", describeLiveDomainsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'scf.DescribeLiveDomains': %w", err)
		}

		if describeLiveDomainsResp.Response == nil {
			break
		}

		for _, domainItem := range describeLiveDomainsResp.Response.Domains {
			domains = append(domains, *domainItem.Domain)
		}

		if len(describeLiveDomainsResp.Response.Domains) < listCustomDomainsLimit {
			break
		}

		listCustomDomainsOffset++
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain string, cloudCertId string) error {
	// 查看云函数自定义域名详情
	// REF: https://cloud.tencent.com/document/api/583/111924
	getCustomDomainReq := tcscf.NewGetCustomDomainRequest()
	getCustomDomainReq.Domain = common.StringPtr(domain)
	getCustomDomainResp, err := d.sdkClient.GetCustomDomainWithContext(ctx, getCustomDomainReq)
	d.logger.Debug("sdk request 'scf.GetCustomDomain'", slog.Any("request", getCustomDomainReq), slog.Any("response", getCustomDomainResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'scf.GetCustomDomain': %w", err)
	} else {
		if getCustomDomainResp.Response.CertConfig != nil && getCustomDomainResp.Response.CertConfig.CertificateId != nil && *getCustomDomainResp.Response.CertConfig.CertificateId == cloudCertId {
			return nil
		}
	}

	// 更新云函数自定义域名
	// REF: https://cloud.tencent.com/document/api/583/111922
	updateCustomDomainReq := tcscf.NewUpdateCustomDomainRequest()
	updateCustomDomainReq.Domain = common.StringPtr(domain)
	updateCustomDomainReq.CertConfig = &tcscf.CertConf{
		CertificateId: common.StringPtr(cloudCertId),
	}
	updateCustomDomainReq.Protocol = getCustomDomainResp.Response.Protocol
	if updateCustomDomainReq.Protocol == nil || *updateCustomDomainReq.Protocol == "HTTP" {
		updateCustomDomainReq.Protocol = common.StringPtr("HTTP&HTTPS")
	}
	updateCustomDomainResp, err := d.sdkClient.UpdateCustomDomainWithContext(ctx, updateCustomDomainReq)
	d.logger.Debug("sdk request 'scf.UpdateCustomDomain'", slog.Any("request", updateCustomDomainReq), slog.Any("response", updateCustomDomainResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'scf.UpdateCustomDomain': %w", err)
	}

	return nil
}

func createSDKClient(secretId, secretKey, endpoint, region string) (*tcscf.Client, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := tcscf.NewClient(credential, region, cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
