package tencentcloudeomakers

import (
	"context"
	"crypto/x509"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	tceo "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/tencentcloud-ssl"
	tceomakersssdk "github.com/certimate-go/certimate/pkg/sdk3rd/tencentcloud/teomakers"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
	xcertkey "github.com/certimate-go/certimate/pkg/utils/cert/key"
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
	// EdgeOne Makers API Token。
	MakersApiToken string `json:"makersApiToken"`
	// EdgeOne Makers 项目 ID。
	MakersProjectId string `json:"makersProjectId"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 加速域名列表（支持泛域名）。
	Domains []string `json:"domains"`
	// 是否启用多证书模式。
	EnableMultipleSSL bool `json:"enableMultipleSSL,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClients *wSDKClients
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

type wSDKClients struct {
	TEO       *tceo.Client
	TEOMakers *tceomakersssdk.Client
}

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	clients, err := createSDKClients(config.SecretId, config.SecretKey, config.Endpoint, config.MakersApiToken)
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
		sdkClients: clients,
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
	if d.config.MakersProjectId == "" {
		return nil, fmt.Errorf("config `makersProjectId` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 获取全部可部署的域名列表
	domainsInMakers, err := d.getAllDomainsInProject(ctx, d.config.MakersProjectId)
	if err != nil {
		return nil, err
	}

	// 获取待部署的域名列表
	var domains []string
	switch d.config.DomainMatchPattern {
	case "", DomainMatchPatternExact:
		{
			if len(d.config.Domains) == 0 {
				return nil, fmt.Errorf("config `domains` is required")
			}

			domains = d.config.Domains
		}

	case DomainMatchPatternWildcard:
		{
			if len(d.config.Domains) == 0 {
				return nil, fmt.Errorf("config `domains` is required")
			}

			domainCandidates := lo.Map(domainsInMakers, func(domainInfo *tceomakersssdk.PagesZoneCustomDomain, _ int) string {
				return lo.FromPtr(domainInfo.Domain)
			})
			domains = lo.Filter(domainCandidates, func(domain string, _ int) bool {
				for _, configDomain := range d.config.Domains {
					if xcerthostname.IsMatch(configDomain, domain) {
						return true
					}
				}
				return false
			})
			if len(domains) == 0 {
				return nil, fmt.Errorf("could not find any domains matched by wildcard")
			}
		}

	case DomainMatchPatternCertSan:
		{
			domainCandidates := lo.Map(domainsInMakers, func(domainInfo *tceomakersssdk.PagesZoneCustomDomain, _ int) string {
				return lo.FromPtr(domainInfo.Domain)
			})
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
		d.logger.Info("no edgeone makers domains to deploy")
	} else {
		d.logger.Info("found edgeone makers domains to deploy", slog.Any("domains", domains))

		// 获取站点 ID
		zoneId := domainsInMakers[0].ZoneId

		// 获取证书列表
		describeHostCertificatesReq := tceo.NewDescribeHostCertificatesRequest()
		describeHostCertificatesReq.ZoneId = zoneId
		describeHostCertificatesResp, err := d.sdkClients.TEO.DescribeHostCertificatesWithContext(ctx, describeHostCertificatesReq)
		d.logger.Debug("sdk request 'teo.DescribeHostCertificates'", slog.Any("request", describeHostCertificatesReq), slog.Any("response", describeHostCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'teo.DescribeHostCertificates': %w", err)
		}

		// 跳过已部署过的域名
		domains = lo.Filter(domains, func(domain string, _ int) bool {
			var deployed bool

			domainInfo, _ := lo.Find(domainsInMakers, func(domainInfo *tceomakersssdk.PagesZoneCustomDomain) bool {
				return domain == lo.FromPtr(domainInfo.Domain)
			})
			if domainInfo != nil && describeHostCertificatesResp.Response != nil {
				deployed = lo.SomeBy(describeHostCertificatesResp.Response.HostCertificates, func(hostInfo *tceo.HostCertificate) bool {
					return domain == lo.FromPtr(hostInfo.Host) &&
						lo.SomeBy(hostInfo.HostCertInfo, func(certInfo *tceo.CertificateInfo) bool {
							return upres.CertId == lo.FromPtr(certInfo.CertId)
						})
				})
			}

			return !deployed
		})

		// 配置域名证书
		// REF: https://cloud.tencent.com/document/api/1552/80764
		requests := make([]*tceo.ModifyHostsCertificateRequest, 0)
		if d.config.EnableMultipleSSL {
			const algRSA = "RSA"
			const algECC = "ECC"

			privkey, err := xcert.ParsePrivateKeyFromPEM(privkeyPEM)
			if err != nil {
				return nil, fmt.Errorf("failed to parse private key: %w", err)
			}

			privkeyAlg, _, _ := xcertkey.GetPrivateKeyAlgorithm(privkey)
			privkeyAlgStr := ""
			switch privkeyAlg {
			case x509.RSA:
				privkeyAlgStr = algRSA
			case x509.ECDSA:
				privkeyAlgStr = algECC
			}

			for _, domain := range domains {
				modifyHostsCertificateReq := tceo.NewModifyHostsCertificateRequest()
				modifyHostsCertificateReq.ZoneId = zoneId
				modifyHostsCertificateReq.Mode = common.StringPtr("sslcert")
				modifyHostsCertificateReq.Hosts = common.StringPtrs([]string{domain})
				modifyHostsCertificateReq.ServerCertInfo = []*tceo.ServerCertInfo{{CertId: common.StringPtr(upres.CertId)}}

				domainInfo, _ := lo.Find(domainsInMakers, func(domainInfo *tceomakersssdk.PagesZoneCustomDomain) bool {
					return domain == lo.FromPtr(domainInfo.Domain)
				})
				if domainInfo != nil && describeHostCertificatesResp.Response != nil {
					for _, hostInfo := range describeHostCertificatesResp.Response.HostCertificates {
						if lo.FromPtr(hostInfo.Host) != domain {
							continue
						}

						for _, certInfo := range hostInfo.HostCertInfo {
							if lo.FromPtr(certInfo.CertId) == upres.CertId {
								continue
							}

							if strings.Split(lo.FromPtr(certInfo.SignAlgo), " ")[0] == privkeyAlgStr {
								continue
							}

							certExpireTime, _ := time.Parse("2006-01-02T15:04:05Z", lo.FromPtr(certInfo.ExpireTime))
							if certExpireTime.Before(time.Now()) {
								continue
							}

							modifyHostsCertificateReq.ServerCertInfo = append(modifyHostsCertificateReq.ServerCertInfo, &tceo.ServerCertInfo{CertId: certInfo.CertId})
						}
					}
				}

				requests = append(requests, modifyHostsCertificateReq)
			}
		} else {
			modifyHostsCertificateReq := tceo.NewModifyHostsCertificateRequest()
			modifyHostsCertificateReq.ZoneId = zoneId
			modifyHostsCertificateReq.Mode = common.StringPtr("sslcert")
			modifyHostsCertificateReq.Hosts = common.StringPtrs(domains)
			modifyHostsCertificateReq.ServerCertInfo = []*tceo.ServerCertInfo{{CertId: common.StringPtr(upres.CertId)}}

			requests = append(requests, modifyHostsCertificateReq)
		}

		if err := xloop.ForRangeAllWithContext(ctx, requests, func(ctx context.Context, modifyHostsCertificateReq *tceo.ModifyHostsCertificateRequest, _ int) error {
			modifyHostsCertificateResp, err := d.sdkClients.TEO.ModifyHostsCertificateWithContext(ctx, modifyHostsCertificateReq)
			d.logger.Debug("sdk request 'teo.ModifyHostsCertificate'", slog.Any("request", modifyHostsCertificateReq), slog.Any("response", modifyHostsCertificateResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'teo.ModifyHostsCertificate': %w", err)
			}

			return nil
		}); err != nil {
			return nil, err
		}
	}

	return &DeployResult{}, nil
}

func (d *Deployer) getAllDomainsInProject(ctx context.Context, makersProjectId string) ([]*tceomakersssdk.PagesZoneCustomDomain, error) {
	// 查询创建过的 Pages 自定义域名列表
	// REF: https://docs.edgeone.site/#/?id=describepageszonecustomdomains
	describeMakersZoneCustomDomainsReq := &tceomakersssdk.DescribePagesZoneCustomDomainsRequest{
		ProjectId: common.StringPtr(makersProjectId),
	}
	describeMakersZoneCustomDomainsResp, err := d.sdkClients.TEOMakers.DescribePagesZoneCustomDomainsWithContext(ctx, describeMakersZoneCustomDomainsReq)
	d.logger.Debug("api request 'teo.makers.DescribePagesZoneCustomDomains'", slog.Any("request", describeMakersZoneCustomDomainsReq), slog.Any("response", describeMakersZoneCustomDomainsResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'teo.makers.DescribePagesZoneCustomDomains': %w", err)
	}

	domains := make([]*tceomakersssdk.PagesZoneCustomDomain, 0)
	for _, domainItem := range describeMakersZoneCustomDomainsResp.Data.Response.PagesDomains {
		if lo.FromPtr(domainItem.Type) == "Custom" && lo.FromPtr(domainItem.ZoneId) != "" {
			domains = append(domains, domainItem)
		}
	}

	return domains, nil
}

func createSDKClients(secretId, secretKey, endpoint, makersApiToken string) (*wSDKClients, error) {
	wsdk := &wSDKClients{}

	{
		credential := common.NewCredential(secretId, secretKey)

		cpf := profile.NewClientProfile()
		if endpoint != "" {
			cpf.HttpProfile.Endpoint = endpoint
		}

		client, err := tceo.NewClient(credential, "", cpf)
		if err != nil {
			return nil, err
		}

		wsdk.TEO = client
	}

	{
		client, err := tceomakersssdk.NewClient(
			tceomakersssdk.WithApiToken(makersApiToken),
		)
		if err != nil {
			return nil, err
		}

		wsdk.TEOMakers = client
	}

	return wsdk, nil
}
