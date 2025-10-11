package tencentcloudeo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcteo "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"

	"github.com/certimate-go/certimate/pkg/core"
	sslmgrsp "github.com/certimate-go/certimate/pkg/core/ssl-manager/providers/tencentcloud-ssl"
)

type SSLDeployerProviderConfig struct {
	// 腾讯云 SecretId。
	SecretId string `json:"secretId"`
	// 腾讯云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 腾讯云接口端点。
	Endpoint string `json:"endpoint,omitempty"`
	// 站点 ID。
	ZoneId string `json:"zoneId"`
	// 域名匹配模式。
	// 零值时默认值 [MatchPatternExact]。
	MatchPattern string `json:"matchPattern,omitempty"`
	// 加速域名列表（支持泛域名）。
	Domains []string `json:"domains"`
}

type SSLDeployerProvider struct {
	config     *SSLDeployerProviderConfig
	logger     *slog.Logger
	sdkClient  *tcteo.Client
	sslManager core.SSLManager
}

var _ core.SSLDeployer = (*SSLDeployerProvider)(nil)

func NewSSLDeployerProvider(config *SSLDeployerProviderConfig) (*SSLDeployerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl deployer provider is nil")
	}

	client, err := createSDKClient(config.SecretId, config.SecretKey, config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("could not create sdk client: %w", err)
	}

	sslmgr, err := sslmgrsp.NewSSLManagerProvider(&sslmgrsp.SSLManagerProviderConfig{
		SecretId:  config.SecretId,
		SecretKey: config.SecretKey,
		Endpoint: lo.
			If(strings.HasSuffix(config.Endpoint, "intl.tencentcloudapi.com"), "ssl.intl.tencentcloudapi.com"). // 国际站使用独立的接口端点
			Else(""),
	})
	if err != nil {
		return nil, fmt.Errorf("could not create ssl manager: %w", err)
	}

	return &SSLDeployerProvider{
		config:     config,
		logger:     slog.Default(),
		sdkClient:  client,
		sslManager: sslmgr,
	}, nil
}

func (d *SSLDeployerProvider) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}

	d.sslManager.SetLogger(logger)
}

func (d *SSLDeployerProvider) Deploy(ctx context.Context, certPEM string, privkeyPEM string) (*core.SSLDeployResult, error) {
	if d.config.ZoneId == "" {
		return nil, errors.New("config `zoneId` is required")
	}
	if len(d.config.Domains) == 0 {
		return nil, errors.New("config `domains` is required")
	}

	// 上传证书
	upres, err := d.sslManager.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	if len(d.config.Domains) == 0 || d.config.Domains[0] == "" {
		return nil, errors.New("config `domains` is required")
	}

	var domains []string
	switch d.config.MatchPattern {
	case "", MatchPatternExact:
		{
			domains = d.config.Domains
		}

	case MatchPatternWildcard:
		{
			domainsInZone, err := d.getDomainsInZone()
			if err != nil {
				return nil, err
			}

			domains = calcWildcardDomainIntersection(d.config.Domains, domainsInZone)
			if len(domains) == 0 {
				return nil, errors.New("no domains matched in wildcard mode")
			}
		}

	default:
		return nil, fmt.Errorf("unsupported match pattern: '%s'", d.config.MatchPattern)
	}

	// 配置域名证书
	// REF: https://cloud.tencent.com/document/api/1552/80764
	modifyHostsCertificateReq := tcteo.NewModifyHostsCertificateRequest()
	modifyHostsCertificateReq.ZoneId = common.StringPtr(d.config.ZoneId)
	modifyHostsCertificateReq.Mode = common.StringPtr("sslcert")
	modifyHostsCertificateReq.Hosts = common.StringPtrs(domains)
	modifyHostsCertificateReq.ServerCertInfo = []*tcteo.ServerCertInfo{{CertId: common.StringPtr(upres.CertId)}}
	modifyHostsCertificateResp, err := d.sdkClient.ModifyHostsCertificate(modifyHostsCertificateReq)
	d.logger.Debug("sdk request 'teo.ModifyHostsCertificate'", slog.Any("request", modifyHostsCertificateReq), slog.Any("response", modifyHostsCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'teo.ModifyHostsCertificate': %w", err)
	}

	return &core.SSLDeployResult{}, nil
}

// getDomainsInZone 获取站点下的所有域名
func (d *SSLDeployerProvider) getDomainsInZone() ([]string, error) {
	var domainsInZone []string
	describeAccelerationDomainsReq := tcteo.NewDescribeAccelerationDomainsRequest()
	describeAccelerationDomainsReq.ZoneId = common.StringPtr(d.config.ZoneId)
	describeAccelerationDomainsResp, err := d.sdkClient.DescribeAccelerationDomains(describeAccelerationDomainsReq)
	d.logger.Debug("sdk request 'teo.DescribeAccelerationDomains'", slog.Any("request", describeAccelerationDomainsReq), slog.Any("response", describeAccelerationDomainsResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'teo.DescribeAccelerationDomains': %w", err)
	}
	if describeAccelerationDomainsResp == nil || describeAccelerationDomainsResp.Response == nil || describeAccelerationDomainsResp.Response.TotalCount == nil {
		return nil, errors.New("unexpected deployment job status")
	}
	for _, accelerationDomain := range describeAccelerationDomainsResp.Response.AccelerationDomains {
		if accelerationDomain == nil || accelerationDomain.DomainName == nil {
			continue
		}
		domainsInZone = append(domainsInZone, *accelerationDomain.DomainName)
	}
	return domainsInZone, nil
}

func calcWildcardDomainIntersection(domains []string, domainsInZone []string) []string {
	var result []string
	for _, domainInZone := range domainsInZone {
		for _, domain := range domains {
			// 精准匹配
			if domainInZone == domain {
				result = append(result, domainInZone)
				break
			}
			// 非泛域名跳过
			if !strings.HasPrefix(domain, "*.") {
				continue
			}
			// 泛域名后缀不匹配
			if !strings.HasSuffix(domainInZone, domain[1:]) {
				continue
			}
			// 如果域名前缀包含点，说明是多级域名，则不匹配
			domainPrefix, _ := strings.CutSuffix(domainInZone, domain[1:])
			if strings.Contains(domainPrefix, ".") {
				continue
			}
			// 泛域名匹配
			result = append(result, domainInZone)
			break
		}
	}
	return result
}

func createSDKClient(secretId, secretKey, endpoint string) (*tcteo.Client, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := tcteo.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
