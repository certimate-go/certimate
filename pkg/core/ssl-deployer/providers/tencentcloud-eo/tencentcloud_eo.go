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
	// 是否部署到站点下的所有域名。
	DeployToAllDomains bool `json:"deployToAllDomains,omitempty"`
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

func (d *SSLDeployerProvider) getAllAccelerationDomains(ctx context.Context, zoneId string) ([]string, error) {
	var allDomains []string
	var offset int64 = 0
	const limit int64 = 100

	for {
		// 获取加速域名列表的请求
		// REF: https://cloud.tencent.com/document/api/1552/80768
		describeReq := tcteo.NewDescribeAccelerationDomainsRequest()
		describeReq.ZoneId = common.StringPtr(zoneId)
		describeReq.Offset = common.Int64Ptr(offset)
		describeReq.Limit = common.Int64Ptr(limit)

		describeResp, err := d.sdkClient.DescribeAccelerationDomains(describeReq)
		d.logger.Debug("sdk request 'teo.DescribeAccelerationDomains'", slog.Any("request", describeReq), slog.Any("response", describeResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'teo.DescribeAccelerationDomains': %w", err)
		}

		// 收集域名
		for _, domain := range describeResp.Response.AccelerationDomains {
			if domain.DomainName != nil {
				allDomains = append(allDomains, *domain.DomainName)
			}
		}

		// 检查是否还有更多数据
		if describeResp.Response.TotalCount == nil || int64(len(allDomains)) >= *describeResp.Response.TotalCount {
			break
		}

		offset += limit
	}

	d.logger.Info("retrieved all acceleration domains", slog.String("zoneId", zoneId), slog.Int("count", len(allDomains)), slog.Any("domains", allDomains))
	return allDomains, nil
}

func (d *SSLDeployerProvider) Deploy(ctx context.Context, certPEM string, privkeyPEM string) (*core.SSLDeployResult, error) {
	if d.config.ZoneId == "" {
		return nil, errors.New("config `zoneId` is required")
	}

	var domainsToUse []string
	
	// 判断是否部署到所有域名
	if d.config.DeployToAllDomains {
		allDomains, err := d.getAllAccelerationDomains(ctx, d.config.ZoneId)
		if err != nil {
			return nil, fmt.Errorf("failed to get all acceleration domains: %w", err)
		}
		if len(allDomains) == 0 {
			return nil, errors.New("no acceleration domains found in the specified zone")
		}
		domainsToUse = allDomains
		d.logger.Info("deploying certificate to all domains", slog.Int("count", len(domainsToUse)))
	} else {
		if len(d.config.Domains) == 0 {
			return nil, errors.New("config `domains` is required when `deployToAllDomains` is false")
		}
		domainsToUse = d.config.Domains
		d.logger.Info("deploying certificate to specified domains", slog.Any("domains", domainsToUse))
	}

	// 上传证书
	upres, err := d.sslManager.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 配置域名证书
	// REF: https://cloud.tencent.com/document/api/1552/80764
	modifyHostsCertificateReq := tcteo.NewModifyHostsCertificateRequest()
	modifyHostsCertificateReq.ZoneId = common.StringPtr(d.config.ZoneId)
	modifyHostsCertificateReq.Mode = common.StringPtr("sslcert")
	modifyHostsCertificateReq.Hosts = common.StringPtrs(domainsToUse)
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
