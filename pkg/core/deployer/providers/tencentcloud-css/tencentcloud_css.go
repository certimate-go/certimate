package tencentcloudcss

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tclive "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/live/v20180801"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/tencentcloud-ssl"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-css/internal"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
)

type DeployerConfig struct {
	// 腾讯云 SecretId。
	SecretId string `json:"secretId"`
	// 腾讯云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 腾讯云接口端点。
	Endpoint string `json:"endpoint,omitempty"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 直播流域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *internal.LiveClient
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

	// 获取待部署的域名列表
	var domains []string
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

			domains = lo.Filter(domainCandidates, func(domain string, _ int) bool {
				return xcerthostname.IsMatchByCertificatePEM(certPEM, domain)
			})
			if len(domains) == 0 {
				return nil, errors.New("could not find any domains matched by certificate")
			}
		}

	default:
		return nil, fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 批量绑定证书对应的播放域名
	// REF: https://cloud.tencent.com/document/api/267/78655
	modifyLiveDomainCertBindingsReq := tclive.NewModifyLiveDomainCertBindingsRequest()
	modifyLiveDomainCertBindingsReq.DomainInfos = lo.Map(domains, func(domain string, _ int) *tclive.LiveCertDomainInfo {
		return &tclive.LiveCertDomainInfo{
			DomainName: common.StringPtr(domain),
			Status:     common.Int64Ptr(1),
		}
	})
	modifyLiveDomainCertBindingsReq.CloudCertId = common.StringPtr(upres.CertId)
	modifyLiveDomainCertBindingsResp, err := d.sdkClient.ModifyLiveDomainCertBindings(modifyLiveDomainCertBindingsReq)
	d.logger.Debug("sdk request 'live.ModifyLiveDomainCertBindings'", slog.Any("request", modifyLiveDomainCertBindingsReq), slog.Any("response", modifyLiveDomainCertBindingsResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'live.ModifyLiveDomainCertBindings': %w", err)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) getAllDomains(ctx context.Context) ([]string, error) {
	domains := make([]string, 0)

	// 查询域名列表
	// REF: https://cloud.tencent.com/document/api/267/33856
	describeLiveDomainsPageNum := 1
	describeLiveDomainsPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		describeLiveDomainsReq := tclive.NewDescribeLiveDomainsRequest()
		describeLiveDomainsReq.DomainStatus = common.Uint64Ptr(1)
		describeLiveDomainsReq.DomainType = common.Uint64Ptr(1)
		describeLiveDomainsReq.PageNum = common.Uint64Ptr(uint64(describeLiveDomainsPageNum))
		describeLiveDomainsReq.PageSize = common.Uint64Ptr(uint64(describeLiveDomainsPageSize))
		describeLiveDomainsResp, err := d.sdkClient.DescribeLiveDomains(describeLiveDomainsReq)
		d.logger.Debug("sdk request 'live.DescribeLiveDomains'", slog.Any("request", describeLiveDomainsReq), slog.Any("response", describeLiveDomainsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'live.DescribeLiveDomains': %w", err)
		}

		if describeLiveDomainsResp.Response == nil {
			break
		}

		for _, domainItem := range describeLiveDomainsResp.Response.DomainList {
			domains = append(domains, *domainItem.Name)
		}

		if len(describeLiveDomainsResp.Response.DomainList) < describeLiveDomainsPageSize {
			break
		}

		describeLiveDomainsPageNum++
	}

	return domains, nil
}

func createSDKClient(secretId, secretKey, endpoint string) (*internal.LiveClient, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := internal.NewLiveClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
