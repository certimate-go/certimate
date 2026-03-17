package baishancdn

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/baishan-cdn"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	baishansdk "github.com/certimate-go/certimate/pkg/sdk3rd/baishan"
)

type DeployerConfig struct {
	// 白山云 API Token。
	ApiToken string `json:"apiToken"`
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// 域名匹配模式。暂时只支持精确匹配。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 加速域名（支持泛域名）。
	Domain string `json:"domain"`
	// 证书 ID。
	// 部署资源类型为 [RESOURCE_TYPE_CERTIFICATE] 时必填。
	CertificateId string `json:"certificateId,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *baishansdk.Client
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ApiToken)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		ApiToken: config.ApiToken,
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
	// 根据部署资源类型决定部署方式
	switch d.config.ResourceType {
	case RESOURCE_TYPE_DOMAIN:
		if err := d.deployToDomain(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	case RESOURCE_TYPE_CERTIFICATE:
		if err := d.deployToCertificate(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported resource type '%s'", d.config.ResourceType)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) deployToDomain(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.Domain == "" {
		return errors.New("config `domain` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 查询域名配置
	// REF: https://portal.baishancloud.com/track/document/api/1/1065
	getDomainConfigReq := &baishansdk.GetDomainConfigRequest{
		Domains: lo.ToPtr(d.config.Domain),
		Config:  lo.ToPtr([]string{"https"}),
	}
	getDomainConfigResp, err := d.sdkClient.GetDomainConfigWithContext(ctx, getDomainConfigReq)
	d.logger.Debug("sdk request 'baishan.GetDomainConfig'", slog.Any("request", getDomainConfigReq), slog.Any("response", getDomainConfigResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'baishan.GetDomainConfig': %w", err)
	} else if len(getDomainConfigResp.Data) == 0 {
		return fmt.Errorf("could not find domain '%s'", d.config.Domain)
	}

	// 设置域名配置
	// REF: https://portal.baishancloud.com/track/document/api/1/1045
	setDomainConfigReq := &baishansdk.SetDomainConfigRequest{
		Domains: lo.ToPtr(d.config.Domain),
		Config: &baishansdk.DomainConfig{
			Https: &baishansdk.DomainConfigHttps{
				CertId:      json.Number(upres.CertId),
				ForceHttps:  getDomainConfigResp.Data[0].Config.Https.ForceHttps,
				EnableHttp2: getDomainConfigResp.Data[0].Config.Https.EnableHttp2,
				EnableOcsp:  getDomainConfigResp.Data[0].Config.Https.EnableOcsp,
			},
		},
	}
	setDomainConfigResp, err := d.sdkClient.SetDomainConfigWithContext(ctx, setDomainConfigReq)
	d.logger.Debug("sdk request 'baishan.SetDomainConfig'", slog.Any("request", setDomainConfigReq), slog.Any("response", setDomainConfigResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'baishan.SetDomainConfig': %w", err)
	}

	return nil
}

func (d *Deployer) deployToCertificate(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.CertificateId == "" {
		return errors.New("config `certificateId` is required")
	}

	// 替换证书
	opres, err := d.sdkCertmgr.Replace(ctx, d.config.CertificateId, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to replace certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate replaced", slog.Any("result", opres))
	}

	return nil
}

func createSDKClient(apiToken string) (*baishansdk.Client, error) {
	return baishansdk.NewClient(apiToken)
}
