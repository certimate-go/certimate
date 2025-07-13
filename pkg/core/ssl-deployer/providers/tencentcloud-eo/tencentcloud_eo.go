package tencentcloudeo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcteo "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"

	"github.com/certimate-go/certimate/pkg/core"
	sslmgrsp "github.com/certimate-go/certimate/pkg/core/ssl-manager/providers/tencentcloud-ssl"
	"github.com/certimate-go/certimate/pkg/utils/ifelse"
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
		Endpoint: ifelse.
			If[string](strings.HasSuffix(strings.TrimSpace(config.Endpoint), "intl.tencentcloudapi.com")).
			Then("ssl.intl.tencentcloudapi.com"). // 国际站使用独立的接口端点
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

	// 若存在泛域名，则拉取Zone下所有域名进行校验
	var updateDomainSet map[string]struct{}
	var domainsInZone []string
	for _, domain := range d.config.Domains {
		if !strings.HasPrefix(domain, "*") {
			updateDomainSet[domain] = struct{}{}
			continue
		}
		// 获取Zone下所有域名
		if domainsInZone == nil {
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
			domainsInZone = make([]string, 0, *describeAccelerationDomainsResp.Response.TotalCount)
			for _, accelerationDomain := range describeAccelerationDomainsResp.Response.AccelerationDomains {
				if accelerationDomain == nil || accelerationDomain.DomainName == nil {
					continue
				}
				domainsInZone = append(domainsInZone, *accelerationDomain.DomainName)
			}
		}
		// 遍历Zone下所有域名，若域名匹配泛域名，则添加到更新域名列表中
		for _, zoneDomain := range domainsInZone {
			if !strings.HasSuffix(zoneDomain, domain[1:]) { // 匹配泛域名
				continue
			}
			domainPrefix, _ := strings.CutSuffix(zoneDomain, domain[1:])
			if strings.Contains(domainPrefix, ".") { // 如果域名前缀包含点，说明是多级域名，则不匹配
				continue
			}
			updateDomainSet[zoneDomain] = struct{}{}
		}
	}

	// 将updateDomainSet转换为slice
	var updateDomains []string
	for domain := range updateDomainSet {
		updateDomains = append(updateDomains, domain)
	}

	// 配置域名证书
	// REF: https://cloud.tencent.com/document/api/1552/80764
	modifyHostsCertificateReq := tcteo.NewModifyHostsCertificateRequest()
	modifyHostsCertificateReq.ZoneId = common.StringPtr(d.config.ZoneId)
	modifyHostsCertificateReq.Mode = common.StringPtr("sslcert")
	modifyHostsCertificateReq.Hosts = common.StringPtrs(updateDomains)
	modifyHostsCertificateReq.ServerCertInfo = []*tcteo.ServerCertInfo{{CertId: common.StringPtr(upres.CertId)}}
	modifyHostsCertificateResp, err := d.sdkClient.ModifyHostsCertificate(modifyHostsCertificateReq)
	d.logger.Debug("sdk request 'teo.ModifyHostsCertificate'", slog.Any("request", modifyHostsCertificateReq), slog.Any("response", modifyHostsCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'teo.ModifyHostsCertificate': %w", err)
	}

	return &core.SSLDeployResult{}, nil
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
