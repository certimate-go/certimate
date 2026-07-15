package tencentcloudvod

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	tcvod "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vod/v20180717"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/tencentcloud-ssl"
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
	// 点播应用 ID。
	SubAppId int64 `json:"subAppId"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 点播加速域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *tcvod.Client
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.SecretId, config.SecretKey, config.Endpoint)
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

	// 获取待部署的 ECDN 实例
	domains := make([]string, 0)
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

			domains = domainCandidates
		}

	default:
		return nil, fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 批量更新域名证书
	if len(domains) == 0 {
		d.logger.Info("no vod domains to deploy")
	} else {
		d.logger.Info("found vod domains to deploy", slog.Any("domains", domains))

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

	// 查询点播域名列表
	// REF: https://cloud.tencent.com/document/api/266/54176
	describeVodDomainsOffset := 0
	describeVodDomainsLimit := 20
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		describeVodDomainsReq := tcvod.NewDescribeVodDomainsRequest()
		describeVodDomainsReq.Offset = common.Uint64Ptr(uint64(describeVodDomainsOffset))
		describeVodDomainsReq.Limit = common.Uint64Ptr(uint64(describeVodDomainsLimit))
		if d.config.SubAppId != 0 {
			describeVodDomainsReq.SubAppId = common.Uint64Ptr(uint64(d.config.SubAppId))
		}
		describeVodDomainsResp, err := d.sdkClient.DescribeVodDomainsWithContext(ctx, describeVodDomainsReq)
		d.logger.Debug("sdk request 'vod.DescribeVodDomains'", slog.Any("request", describeVodDomainsReq), slog.Any("response", describeVodDomainsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'vod.DescribeVodDomains': %w", err)
		}

		if describeVodDomainsResp.Response == nil {
			break
		}

		ignoredStatuses := []string{"Locked"}
		for _, domainItem := range describeVodDomainsResp.Response.DomainSet {
			if lo.Contains(ignoredStatuses, *domainItem.DeployStatus) {
				continue
			}

			domains = append(domains, *domainItem.Domain)
		}

		if len(describeVodDomainsResp.Response.DomainSet) < describeVodDomainsLimit {
			break
		}

		describeVodDomainsOffset += describeVodDomainsLimit
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain string, cloudCertId string) error {
	// 设置点播域名 HTTPS 证书
	// REF: https://cloud.tencent.com/document/api/266/102015
	setVodDomainCertificateReq := tcvod.NewSetVodDomainCertificateRequest()
	setVodDomainCertificateReq.Domain = common.StringPtr(domain)
	setVodDomainCertificateReq.Operation = common.StringPtr("Set")
	setVodDomainCertificateReq.CertID = common.StringPtr(cloudCertId)
	if d.config.SubAppId != 0 {
		setVodDomainCertificateReq.SubAppId = common.Uint64Ptr(uint64(d.config.SubAppId))
	}
	setVodDomainCertificateResp, err := d.sdkClient.SetVodDomainCertificateWithContext(ctx, setVodDomainCertificateReq)
	d.logger.Debug("sdk request 'vod.SetVodDomainCertificate'", slog.Any("request", setVodDomainCertificateReq), slog.Any("response", setVodDomainCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'vod.SetVodDomainCertificate': %w", err)
	}

	return nil
}

func createSDKClient(secretId, secretKey, endpoint string) (*tcvod.Client, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := tcvod.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
