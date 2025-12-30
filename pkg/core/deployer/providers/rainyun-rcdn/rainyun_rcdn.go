package rainyunrcdn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/rainyun-sslcenter"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	rainyunsdk "github.com/certimate-go/certimate/pkg/sdk3rd/rainyun"
)

type DeployerConfig struct {
	// 雨云 API 密钥。
	ApiKey string `json:"apiKey"`
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// RCDN 实例 ID。
	// 部署资源类型为 [RESOURCE_TYPE_DOMAIN] 时必填。
	InstanceId int64 `json:"instanceId"`
	// 域名匹配模式。暂时只支持精确匹配。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 加速域名（支持泛域名）。
	Domain string `json:"domain"`
	// 证书 ID。
	// 部署资源类型为 [RESOURCE_TYPE_CERTIFICATE] 时必填。
	CertificateId int64 `json:"certificateId,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *rainyunsdk.Client
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ApiKey)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		ApiKey: config.ApiKey,
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
	if d.config.InstanceId == 0 {
		return fmt.Errorf("config `instanceId` is required")
	}
	if d.config.Domain == "" {
		return fmt.Errorf("config `domain` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// RCDN SSL 绑定域名
	// REF: https://apifox.com/apidoc/shared/a4595cc8-44c5-4678-a2a3-eed7738dab03/api-184214120
	certId, _ := strconv.ParseInt(upres.CertId, 10, 64)
	rcdnInstanceSslBindReq := &rainyunsdk.RcdnInstanceSslBindRequest{
		CertId:  certId,
		Domains: []string{d.config.Domain},
	}
	rcdnInstanceSslBindResp, err := d.sdkClient.RcdnInstanceSslBind(d.config.InstanceId, rcdnInstanceSslBindReq)
	d.logger.Debug("sdk request 'rcdn.InstanceSslBind'", slog.Int64("instanceId", d.config.InstanceId), slog.Any("request", rcdnInstanceSslBindReq), slog.Any("response", rcdnInstanceSslBindResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'rcdn.InstanceSslBind': %w", err)
	}

	return nil
}

func (d *Deployer) deployToCertificate(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.CertificateId == 0 {
		return errors.New("config `certificateId` is required")
	}

	// 替换证书
	opres, err := d.sdkCertmgr.Replace(ctx, strconv.FormatInt(d.config.CertificateId, 10), certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to replace certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate replaced", slog.Any("result", opres))
	}

	return nil
}

func createSDKClient(apiKey string) (*rainyunsdk.Client, error) {
	return rainyunsdk.NewClient(apiKey)
}
