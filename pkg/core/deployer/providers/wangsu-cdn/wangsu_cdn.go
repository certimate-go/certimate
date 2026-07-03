package wangsucdn

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/wangsu-certificate"
	wangsucdn "github.com/certimate-go/certimate/pkg/sdk3rd/wangsu/cdn"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// 网宿云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 网宿云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 域名匹配模式。暂时只支持精确匹配。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 加速域名数组（支持泛域名）。
	Domains []string `json:"domains"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *wangsucdn.Client
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
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

	// 获取待部署的域名列表
	domains := make([]string, 0)
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if len(d.config.Domains) == 0 {
				return nil, fmt.Errorf("config `domains` is required")
			}

			domains = lo.Map(d.config.Domains, func(domain string, _ int) string {
				return normalizeDomain(domain)
			})
		}

	default:
		return nil, fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 批量修改域名证书配置
	// REF: https://www.wangsu.com/document/api-doc/37447
	certId, _ := strconv.ParseInt(upres.CertId, 10, 64)
	batchUpdateCertificateConfigReq := &wangsucdn.BatchUpdateCertificateConfigRequest{
		CertificateId: certId,
		DomainNames:   domains,
	}
	batchUpdateCertificateConfigResp, err := d.sdkClient.BatchUpdateCertificateConfigWithContext(ctx, batchUpdateCertificateConfigReq)
	d.logger.Debug("sdk request 'cdn.BatchUpdateCertificateConfig'", slog.Any("request", batchUpdateCertificateConfigReq), slog.Any("response", batchUpdateCertificateConfigResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cdn.BatchUpdateCertificateConfig': %w", err)
	}

	return &DeployResult{}, nil
}

func createSDKClient(accessKeyId, accessKeySecret string) (*wangsucdn.Client, error) {
	client, err := wangsucdn.NewClient(
		wangsucdn.WithAkSk(accessKeyId, accessKeySecret),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func normalizeDomain(domain string) string {
	// "*.example.com" → ".example.com"，适配网宿云 CDN 的泛域名参数要求
	if strings.HasPrefix(domain, "*.") {
		return strings.TrimPrefix(domain, "*")
	}
	return domain
}
