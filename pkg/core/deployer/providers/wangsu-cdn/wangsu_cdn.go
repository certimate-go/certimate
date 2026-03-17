package wangsucdn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/wangsu-certificate"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	wangsusdk "github.com/certimate-go/certimate/pkg/sdk3rd/wangsu/cdn"
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
	sdkClient  *wangsusdk.Client
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
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
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
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
				return nil, errors.New("config `domains` is required")
			}

			// "*.example.com" → ".example.com"，适配网宿云 CDN 要求的泛域名格式
			domains = lo.Map(d.config.Domains, func(domain string, _ int) string {
				return strings.TrimPrefix(domain, "*")
			})
		}

	default:
		return nil, fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 批量修改域名证书配置
	// REF: https://www.wangsu.com/document/api-doc/37447
	certId, _ := strconv.ParseInt(upres.CertId, 10, 64)
	batchUpdateCertificateConfigReq := &wangsusdk.BatchUpdateCertificateConfigRequest{
		CertificateId: certId,
		DomainNames:   domains,
	}
	batchUpdateCertificateConfigResp, err := d.sdkClient.BatchUpdateCertificateConfigWithContext(ctx, batchUpdateCertificateConfigReq)
	d.logger.Debug("sdk request 'cdn.BatchUpdateCertificateConfig'", slog.Any("request", batchUpdateCertificateConfigReq), slog.Any("response", batchUpdateCertificateConfigResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cdn.BatchUpdateCertificateConfig': %w", err)
	}

	return &deployer.DeployResult{}, nil
}

func createSDKClient(accessKeyId, accessKeySecret string) (*wangsusdk.Client, error) {
	return wangsusdk.NewClient(accessKeyId, accessKeySecret)
}
