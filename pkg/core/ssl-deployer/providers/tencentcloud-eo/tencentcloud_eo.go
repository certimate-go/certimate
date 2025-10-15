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
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
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

	var domains []string
	switch d.config.MatchPattern {
	case "", MatchPatternExact:
		{
			if len(d.config.Domains) == 0 {
				return nil, errors.New("config `domains` is required")
			}

			domains = d.config.Domains
		}

	case MatchPatternWildcard:
		{
			if len(d.config.Domains) == 0 {
				return nil, errors.New("config `domains` is required")
			}

			domainsInZone, err := d.getDomainsInZone(ctx, d.config.ZoneId)
			if err != nil {
				return nil, err
			}

			domains = lo.Filter(domainsInZone, func(domain string, _ int) bool {
				for _, configDomain := range d.config.Domains {
					if xcert.MatchHostname(configDomain, domain) {
						return true
					}
				}
				return false
			})

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

func (d *SSLDeployerProvider) getDomainsInZone(ctx context.Context, zoneId string) ([]string, error) {
	var domainsInZone []string

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

		for _, accelerationDomain := range describeAccelerationDomainsResp.Response.AccelerationDomains {
			if accelerationDomain == nil || accelerationDomain.DomainName == nil {
				continue
			}
			domainsInZone = append(domainsInZone, *accelerationDomain.DomainName)
		}

		if len(describeAccelerationDomainsResp.Response.AccelerationDomains) < pageSize {
			break
		}
	}

	return domainsInZone, nil
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
