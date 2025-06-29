package ctcccloudao

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/certimate-go/certimate/pkg/core"
	sslmgrsp "github.com/certimate-go/certimate/pkg/core/ssl-manager/providers/ctcccloud-ao"
	ctyunao "github.com/certimate-go/certimate/pkg/sdk3rd/ctyun/ao"
	xslices "github.com/certimate-go/certimate/pkg/utils/slices"
	xtypes "github.com/certimate-go/certimate/pkg/utils/types"
)

type SSLDeployerProviderConfig struct {
	// 天翼云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 天翼云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 加速域名（支持泛域名）。
	Domain string `json:"domain"`
}

type SSLDeployerProvider struct {
	config     *SSLDeployerProviderConfig
	logger     *slog.Logger
	sdkClient  *ctyunao.Client
	sslManager core.SSLManager
}

var _ core.SSLDeployer = (*SSLDeployerProvider)(nil)

func NewSSLDeployerProvider(config *SSLDeployerProviderConfig) (*SSLDeployerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey)
	if err != nil {
		return nil, fmt.Errorf("could not create sdk client: %w", err)
	}

	sslmgr, err := sslmgrsp.NewSSLManagerProvider(&sslmgrsp.SSLManagerProviderConfig{
		AccessKeyId:     config.AccessKeyId,
		SecretAccessKey: config.SecretAccessKey,
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
}

func (d *SSLDeployerProvider) Deploy(ctx context.Context, certPEM string, privkeyPEM string) (*core.SSLDeployResult, error) {
	if d.config.Domain == "" {
		return nil, errors.New("config `domain` is required")
	}

	// 上传证书
	upres, err := d.sslManager.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 域名基础及加速配置查询
	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=113&api=13412&data=174&isNormal=1&vid=167
	getDomainConfigReq := &ctyunao.GetDomainConfigRequest{
		Domain:      xtypes.ToPtr(d.config.Domain),
		ProductCode: xtypes.ToPtr("020"),
	}
	getDomainConfigResp, err := d.sdkClient.GetDomainConfig(getDomainConfigReq)
	d.logger.Debug("sdk request 'cdn.GetDomainConfig'", slog.Any("request", getDomainConfigReq), slog.Any("response", getDomainConfigResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cdn.GetDomainConfig': %w", err)
	}

	// 域名基础及加速配置修改
	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=113&api=13413&data=174&isNormal=1&vid=167
	modifyDomainConfigReq := &ctyunao.ModifyDomainConfigRequest{
		Domain:      xtypes.ToPtr(d.config.Domain),
		ProductCode: xtypes.ToPtr(getDomainConfigResp.ReturnObj.ProductCode),
		Origin: xslices.Map(getDomainConfigResp.ReturnObj.Origin, func(item *ctyunao.DomainOriginConfigWithWeight) *ctyunao.DomainOriginConfig {
			weight := item.Weight
			if weight == 0 {
				weight = 1
			}
			return &ctyunao.DomainOriginConfig{
				Origin: item.Origin,
				Role:   item.Role,
				Weight: strconv.Itoa(int(weight)),
			}
		}),
		HttpsStatus: xtypes.ToPtr("on"),
		CertName:    xtypes.ToPtr(upres.CertName),
	}
	modifyDomainConfigResp, err := d.sdkClient.ModifyDomainConfig(modifyDomainConfigReq)
	d.logger.Debug("sdk request 'cdn.ModifyDomainConfig'", slog.Any("request", modifyDomainConfigReq), slog.Any("response", modifyDomainConfigResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cdn.ModifyDomainConfig': %w", err)
	}

	return &core.SSLDeployResult{}, nil
}

func createSDKClient(accessKeyId, secretAccessKey string) (*ctyunao.Client, error) {
	return ctyunao.NewClient(accessKeyId, secretAccessKey)
}
