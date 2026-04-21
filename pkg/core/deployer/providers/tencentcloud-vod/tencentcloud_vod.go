package tencentcloudvod

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcvod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vod/v20180717"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/tencentcloud-ssl"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-vod/internal"
)

type DeployerConfig struct {
	// 腾讯云 SecretId。
	SecretId string `json:"secretId"`
	// 腾讯云 SecretKey。
	SecretKey string `json:"secretKey"`
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
	sdkClient  *internal.VodClient
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.SecretId, config.SecretKey, config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		SecretId:  config.SecretId,
		SecretKey: config.SecretKey,
		Endpoint: lo.
			If(strings.HasSuffix(config.Endpoint, "intl.tencentcloudapi.com"), "ssl.intl.tencentcloudapi.com"). // 国际站使用独立的接口端点
			Else(""),
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

	// 获取待部署的 ECDN 实例
	domains := make([]string, 0)
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if d.config.Domain == "" {
				return nil, errors.New("config `domain` is required")
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

	// 遍历更新域名证书
	if len(domains) == 0 {
		d.logger.Info("no vod domains to deploy")
	} else {
		d.logger.Info("found vod domains to deploy", slog.Any("domains", domains))
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
		describeVodDomainsResp, err := d.sdkClient.DescribeVodDomains(describeVodDomainsReq)
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
	setVodDomainCertificateResp, err := d.sdkClient.SetVodDomainCertificate(setVodDomainCertificateReq)
	d.logger.Debug("sdk request 'vod.SetVodDomainCertificate'", slog.Any("request", setVodDomainCertificateReq), slog.Any("response", setVodDomainCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'vod.SetVodDomainCertificate': %w", err)
	}

	return nil
}

func createSDKClient(secretId, secretKey, endpoint string) (*internal.VodClient, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := internal.NewVodClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
