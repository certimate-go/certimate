package tencentcloudeomakers

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/tencentcloud-ssl"
	tceo "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"
	teom "github.com/certimate-go/certimate/pkg/sdk3rd/tencentcloud/eomakers"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
	xcertkey "github.com/certimate-go/certimate/pkg/utils/cert/key"
	xtencentcloud "github.com/certimate-go/certimate/pkg/utils/third-party/tencentcloud"
	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// Tencent Cloud SecretId
	SecretId string `json:"secretId"`
	// Tencent Cloud SecretKey
	SecretKey string `json:"secretKey"`
	// Tencent Cloud EdgeOne Makers APIToken
	APIToken string `json:"apiToken"`
	// Tencent Cloud ProjectId
	ProjectId int64 `json:"projectId,omitempty"`
	// Tencent Cloud Endpoint
	Endpoint string `json:"endpoint,omitempty"`
	// Tencent Cloud MakersId
	MakersId string `json:"makersId,omitempty"`
	// DomainMatchPattern is the match mode of the domain
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// Domains is the list of acceleration domains
	Domains []string `json:"domains"`
	// EnableMultipleSSL whether the user has enabled multiple certificates mode
	EnableMultipleSSL bool `json:"enableMultipleSSL,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	apiClient  *teom.Client
	sdkClient  *tceo.Client
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	apiClient := teom.NewClient(config.APIToken)

	sdkClient, err := createSDKClient(config.SecretId, config.SecretKey, config.Endpoint)
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
		apiClient:  apiClient,
		sdkClient:  sdkClient,
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
	if d.config.APIToken == "" {
		return nil, fmt.Errorf("config `apiToken` is required")
	}

	if d.config.MakersId == "" {
		return nil, fmt.Errorf("config `makersId` is required")
	}

	// Upload certificate
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// Get all deployable domains
	domainsInMakers, err := d.getAllDomainsInMakers(ctx, d.config.MakersId)
	if err != nil {
		return nil, err
	}
	if len(domainsInMakers) == 0 {
		return nil, fmt.Errorf("could not find any custom domains in edgeone makers project")
	}

	// Get domains ready to deploy
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

			domainCandidates := lo.Map(domainsInMakers, func(domainInfo *teom.DescribePagesZoneCustomDomains, _ int) string {
				return domainInfo.Domain
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
			domainCandidates := lo.Map(domainsInMakers, func(domainInfo *teom.DescribePagesZoneCustomDomains, _ int) string {
				return domainInfo.Domain
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

	// Batch deploy domains
	if len(domains) == 0 {
		d.logger.Info("no edgeone makers domains to deploy")
	} else {
		d.logger.Info("found edgeone makers domains to deploy", slog.Any("domains", domains))

		// In Tencent Cloud Makers, all domains belong to a same special zone
		// You can find this at the console: https://console.cloud.tencent.com/edgeone/makers
		zoneId := domainsInMakers[0].ZoneId

		// Get host certificates
		describeHostCertificatesReq := tceo.NewDescribeHostCertificatesRequest()
		describeHostCertificatesReq.ZoneId = zoneId
		describeHostCertificatesResp, err := d.sdkClient.DescribeHostCertificatesWithContext(
			ctx, describeHostCertificatesReq,
		)
		d.logger.Debug(
			"sdk request 'teo.DescribeHostCertificates'",
			slog.Any("request", describeHostCertificatesReq),
			slog.Any("response", describeHostCertificatesResp),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'teo.DescribeHostCertificates': %w", err)
		}

		// Skip deployed domains
		domains = lo.Filter(domains, func(domain string, _ int) bool {
			var deployed bool

			domainInfo, _ := lo.Find(domainsInMakers, func(domainInfo *teom.DescribePagesZoneCustomDomains) bool {
				return domain == domainInfo.Domain
			})
			if domainInfo != nil && len(describeHostCertificatesResp.Response.HostCertificates) != 0 {
				var certList []tceo.HostCertInfoItem
				for _, v := range describeHostCertificatesResp.Response.HostCertificates {
					if v.Host == domainInfo.Domain {
						certList = v.HostCertInfo
						break
					}
				}

				deployed = lo.SomeBy(certList, func(certInfo tceo.HostCertInfoItem) bool {
					return upres.CertId == certInfo.CertId
				})
			}

			return !deployed
		})

		// Configure certificate of domains
		// REF: https://cloud.tencent.com/document/api/1552/80764
		// REF: https://docs.edgeone.site/#/?id=modifyhostscertificate
		modifyHostsCertificateReqs := make([]*tceo.ModifyHostsCertificateRequest, 0)

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
				modifyHostsCertificateReq.ZoneId = common.StringPtr(zoneId)
				modifyHostsCertificateReq.Mode = common.StringPtr("sslcert")
				modifyHostsCertificateReq.Hosts = common.StringPtrs([]string{domain})
				modifyHostsCertificateReq.ServerCertInfo = []*tceo.ServerCertInfo{{CertId: common.StringPtr(upres.CertId)}}

				domainInfo, _ := lo.Find(domainsInMakers, func(domainInfo *teom.DescribePagesZoneCustomDomains) bool {
					return domain == domainInfo.Domain
				})
				if domainInfo != nil && len(describeHostCertificatesResp.Response.HostCertificates) != 0 {
					var certList []tceo.HostCertInfoItem
					for _, v := range describeHostCertificatesResp.Response.HostCertificates {
						if v.Host == domainInfo.Domain {
							certList = v.HostCertInfo
							break
						}
					}

					for _, certInfo := range certList {
						if certInfo.CertId == upres.CertId {
							continue
						}

						if strings.Split(certInfo.SignAlgo, " ")[0] == privkeyAlgStr {
							continue
						}

						certExpireTime, _ := time.Parse("2006-01-02T15:04:05Z", certInfo.ExpireTime)
						if certExpireTime.Before(time.Now()) {
							continue
						}

						modifyHostsCertificateReq.ServerCertInfo = append(
							modifyHostsCertificateReq.ServerCertInfo, &tceo.ServerCertInfo{CertId: common.StringPtr(certInfo.CertId)},
						)
					}
				}

				modifyHostsCertificateReqs = append(modifyHostsCertificateReqs, modifyHostsCertificateReq)
			}
		} else {
			modifyHostsCertificateReq := tceo.NewModifyHostsCertificateRequest()
			modifyHostsCertificateReq.ZoneId = common.StringPtr(zoneId)
			modifyHostsCertificateReq.Mode = common.StringPtr("sslcert")
			modifyHostsCertificateReq.Hosts = common.StringPtrs(domains)
			modifyHostsCertificateReq.ServerCertInfo = []*tceo.ServerCertInfo{{CertId: common.StringPtr(upres.CertId)}}

			modifyHostsCertificateReqs = append(modifyHostsCertificateReqs, modifyHostsCertificateReq)
		}

		var errs []error
		for _, modifyHostsCertificateReq := range modifyHostsCertificateReqs {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				modifyHostsCertificateResp, err := d.sdkClient.ModifyHostsCertificateWithContext(ctx, modifyHostsCertificateReq)
				d.logger.Debug("sdk request 'teo.ModifyHostsCertificate'",
					slog.Any("request", modifyHostsCertificateReq),
					slog.Any("response", modifyHostsCertificateResp),
				)
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

	return &DeployResult{}, nil
}

func (d *Deployer) getAllDomainsInMakers(ctx context.Context, makersId string) (
	[]*teom.DescribePagesZoneCustomDomains, error,
) {
	// Get list of custom maker domains
	// REF: https://docs.edgeone.site/#/?id=describepageszonecustomdomains
	describeMakersZoneCustomDomainsReq := teom.NewDescribeMakersZoneCustomDomainsReq()
	describeMakersZoneCustomDomainsReq.ProjectId = makersId
	describeMakersZoneCustomDomainsResp, err := d.apiClient.DescribeMakersZoneCustomDomains(
		ctx, describeMakersZoneCustomDomainsReq,
	)
	d.logger.Debug(
		"api request 'eomakers.DescribeMakersZoneCustomDomains'",
		slog.Any("request", describeMakersZoneCustomDomainsReq),
		slog.Any("response", describeMakersZoneCustomDomainsResp),
	)
	if err != nil {
		return nil, err
	}

	result := make([]*teom.DescribePagesZoneCustomDomains, 0)
	for i := range describeMakersZoneCustomDomainsResp.Data.Response.PagesDomains {
		data := &describeMakersZoneCustomDomainsResp.Data.Response.PagesDomains[i]
		if data.Type == "Custom" && data.ZoneId != "" {
			result = append(result, data)
		}
	}

	return result, nil
}

func createSDKClient(secretId, secretKey, endpoint string) (*tceo.Client, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := tceo.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
