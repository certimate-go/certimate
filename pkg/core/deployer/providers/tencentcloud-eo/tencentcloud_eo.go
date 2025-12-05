package tencentcloudeo

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcteo "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/tencentcloud-ssl"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-eo/internal"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
	xcryptokey "github.com/certimate-go/certimate/pkg/utils/crypto/key"
)

type DeployerConfig struct {
	// 腾讯云 SecretId。
	SecretId string `json:"secretId"`
	// 腾讯云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 腾讯云接口端点。
	Endpoint string `json:"endpoint,omitempty"`
	// 站点 ID。
	ZoneId string `json:"zoneId"`
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
	sdkClient  *internal.TeoClient
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	if d.config.ZoneId == "" {
		return nil, errors.New("config `zoneId` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 获取全部可部署的域名信息
	domainsInZone, err := d.getAllDomainsInZone(ctx, d.config.ZoneId)
	if err != nil {
		return nil, err
	}

	// 获取待部署的域名列表
	var domains []string
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if len(d.config.Domains) == 0 {
				return nil, errors.New("config `domains` is required")
			}

			domains = d.config.Domains
		}

	case DOMAIN_MATCH_PATTERN_WILDCARD:
		{
			if len(d.config.Domains) == 0 {
				return nil, errors.New("config `domains` is required")
			}

			domainCandidates := lo.Map(domainsInZone, func(domainInfo *tcteo.AccelerationDomain, _ int) string {
				return lo.FromPtr(domainInfo.DomainName)
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
				return nil, errors.New("could not find any domains matched by wildcard")
			}
		}

	case DOMAIN_MATCH_PATTERN_CERTSAN:
		{
			certX509, err := xcert.ParseCertificateFromPEM(certPEM)
			if err != nil {
				return nil, err
			}

			domainCandidates := lo.Map(domainsInZone, func(domainInfo *tcteo.AccelerationDomain, _ int) string {
				return lo.FromPtr(domainInfo.DomainName)
			})
			domains = lo.Filter(domainCandidates, func(domain string, _ int) bool {
				return certX509.VerifyHostname(domain) == nil
			})
			if len(domains) == 0 {
				return nil, errors.New("could not find any domains matched by certificate")
			}
		}

	default:
		return nil, fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 跳过已部署过的域名
	domains = lo.Filter(domains, func(domain string, _ int) bool {
		var deployed bool

		domainInfo, _ := lo.Find(domainsInZone, func(domainInfo *tcteo.AccelerationDomain) bool {
			return domain == lo.FromPtr(domainInfo.DomainName)
		})
		if domainInfo != nil && domainInfo.Certificate != nil {
			deployed = lo.ContainsBy(domainInfo.Certificate.List, func(certInfo *tcteo.CertificateInfo) bool {
				return upres.CertId == lo.FromPtr(certInfo.CertId)
			})
		}

		return !deployed
	})

	// 批量更新域名证书
	if len(domains) == 0 {
		d.logger.Info("no edgeone domains to deploy")
	} else {
		d.logger.Info("found edgeone domains to deploy", slog.Any("domains", domains))

		// 配置域名证书
		// REF: https://cloud.tencent.com/document/api/1552/80764
		modifyHostsCertificateReqs := make([]*tcteo.ModifyHostsCertificateRequest, 0)

		if d.config.EnableMultipleSSL {
			const algRSA = "RSA"
			const algECC = "ECC"

			privkey, err := xcert.ParsePrivateKeyFromPEM(privkeyPEM)
			if err != nil {
				return nil, fmt.Errorf("failed to parse private key: %w", err)
			}

			privkeyAlg, _, _ := xcryptokey.GetPrivateKeyAlgorithm(privkey)
			privkeyAlgStr := ""
			switch privkeyAlg {
			case x509.RSA:
				privkeyAlgStr = algRSA
			case x509.ECDSA:
				privkeyAlgStr = algECC
			}

			for _, domain := range domains {
				modifyHostsCertificateReq := tcteo.NewModifyHostsCertificateRequest()
				modifyHostsCertificateReq.ZoneId = common.StringPtr(d.config.ZoneId)
				modifyHostsCertificateReq.Mode = common.StringPtr("sslcert")
				modifyHostsCertificateReq.Hosts = common.StringPtrs([]string{domain})
				modifyHostsCertificateReq.ServerCertInfo = []*tcteo.ServerCertInfo{{CertId: common.StringPtr(upres.CertId)}}

				domainInfo, _ := lo.Find(domainsInZone, func(domainInfo *tcteo.AccelerationDomain) bool {
					return domain == lo.FromPtr(domainInfo.DomainName)
				})
				if domainInfo != nil && domainInfo.Certificate != nil {
					for _, certInfo := range domainInfo.Certificate.List {
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

						modifyHostsCertificateReq.ServerCertInfo = append(modifyHostsCertificateReq.ServerCertInfo, &tcteo.ServerCertInfo{CertId: certInfo.CertId})
					}
				}

				modifyHostsCertificateReqs = append(modifyHostsCertificateReqs, modifyHostsCertificateReq)
			}
		} else {
			modifyHostsCertificateReq := tcteo.NewModifyHostsCertificateRequest()
			modifyHostsCertificateReq.ZoneId = common.StringPtr(d.config.ZoneId)
			modifyHostsCertificateReq.Mode = common.StringPtr("sslcert")
			modifyHostsCertificateReq.Hosts = common.StringPtrs(domains)
			modifyHostsCertificateReq.ServerCertInfo = []*tcteo.ServerCertInfo{{CertId: common.StringPtr(upres.CertId)}}

			modifyHostsCertificateReqs = append(modifyHostsCertificateReqs, modifyHostsCertificateReq)
		}

		var errs []error
		for _, modifyHostsCertificateReq := range modifyHostsCertificateReqs {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				modifyHostsCertificateResp, err := d.sdkClient.ModifyHostsCertificate(modifyHostsCertificateReq)
				d.logger.Debug("sdk request 'teo.ModifyHostsCertificate'", slog.Any("request", modifyHostsCertificateReq), slog.Any("response", modifyHostsCertificateResp))
				if err != nil {
					err = fmt.Errorf("failed to execute sdk request 'teo.ModifyHostsCertificate': %w", err)
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

func (d *Deployer) getAllDomainsInZone(ctx context.Context, zoneId string) ([]*tcteo.AccelerationDomain, error) {
	var domainsInZone []*tcteo.AccelerationDomain

	const pageSize = 200
	for offset := 0; ; offset += pageSize {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// 查询加速域名列表
		// REF: https://cloud.tencent.com/document/api/1552/86336
		describeAccelerationDomainsReq := tcteo.NewDescribeAccelerationDomainsRequest()
		describeAccelerationDomainsReq.Limit = common.Int64Ptr(pageSize)
		describeAccelerationDomainsReq.Offset = common.Int64Ptr(int64(offset))
		describeAccelerationDomainsReq.ZoneId = common.StringPtr(zoneId)
		describeAccelerationDomainsResp, err := d.sdkClient.DescribeAccelerationDomains(describeAccelerationDomainsReq)
		d.logger.Debug("sdk request 'teo.DescribeAccelerationDomains'", slog.Any("request", describeAccelerationDomainsReq), slog.Any("response", describeAccelerationDomainsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'teo.DescribeAccelerationDomains': %w", err)
		}

		ignoredStatuses := []string{"offline", "forbidden", "init"}
		for _, domainItem := range describeAccelerationDomainsResp.Response.AccelerationDomains {
			if lo.Contains(ignoredStatuses, lo.FromPtr(domainItem.DomainStatus)) {
				continue
			}

			domainsInZone = append(domainsInZone, domainItem)
		}

		if len(describeAccelerationDomainsResp.Response.AccelerationDomains) < pageSize {
			break
		}
	}

	return domainsInZone, nil
}

func createSDKClient(secretId, secretKey, endpoint string) (*internal.TeoClient, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := internal.NewTeoClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
