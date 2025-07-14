package huaweicloudcdn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	hccdn "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdn/v2"
	hccdnmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdn/v2/model"
	hccdnregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdn/v2/region"

	"github.com/certimate-go/certimate/pkg/core"
	sslmgrsp "github.com/certimate-go/certimate/pkg/core/ssl-manager/providers/huaweicloud-scm"
	xtypes "github.com/certimate-go/certimate/pkg/utils/types"
)

type SSLDeployerProviderConfig struct {
	// 华为云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 华为云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 华为云企业项目 ID。
	EnterpriseProjectId string `json:"enterpriseProjectId,omitempty"`
	// 华为云区域。
	Region string `json:"region"`
	// 加速域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type SSLDeployerProvider struct {
	config     *SSLDeployerProviderConfig
	logger     *slog.Logger
	sdkClient  *hccdn.CdnClient
	sslManager core.SSLManager
}

var _ core.SSLDeployer = (*SSLDeployerProvider)(nil)

func NewSSLDeployerProvider(config *SSLDeployerProviderConfig) (*SSLDeployerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl deployer provider is nil")
	}

	client, err := createSDKClient(
		config.AccessKeyId,
		config.SecretAccessKey,
		config.Region,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create sdk client: %w", err)
	}

	sslmgr, err := sslmgrsp.NewSSLManagerProvider(&sslmgrsp.SSLManagerProviderConfig{
		AccessKeyId:         config.AccessKeyId,
		SecretAccessKey:     config.SecretAccessKey,
		EnterpriseProjectId: config.EnterpriseProjectId,
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
	if d.config.Domain == "" {
		return nil, fmt.Errorf("config `domain` is required")
	}

	// 上传证书
	upres, err := d.sslManager.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 查询加速域名配置
	// REF: https://support.huaweicloud.com/api-cdn/ShowDomainFullConfig.html
	showDomainFullConfigReq := &hccdnmodel.ShowDomainFullConfigRequest{
		EnterpriseProjectId: xtypes.ToPtrOrZeroNil(d.config.EnterpriseProjectId),
		DomainName:          d.config.Domain,
	}
	showDomainFullConfigResp, err := d.sdkClient.ShowDomainFullConfig(showDomainFullConfigReq)
	d.logger.Debug("sdk request 'cdn.ShowDomainFullConfig'", slog.Any("request", showDomainFullConfigReq), slog.Any("response", showDomainFullConfigResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cdn.ShowDomainFullConfig': %w", err)
	}

	// 更新加速域名配置
	// REF: https://support.huaweicloud.com/api-cdn/UpdateDomainMultiCertificates.html
	// REF: https://support.huaweicloud.com/usermanual-cdn/cdn_01_0306.html
	updateDomainMultiCertificatesReqBodyContent := &hccdnmodel.UpdateDomainMultiCertificatesRequestBodyContent{}
	updateDomainMultiCertificatesReqBodyContent.DomainName = d.config.Domain
	updateDomainMultiCertificatesReqBodyContent.HttpsSwitch = 1
	updateDomainMultiCertificatesReqBodyContent.CertificateType = xtypes.ToPtr(int32(2))
	updateDomainMultiCertificatesReqBodyContent.ScmCertificateId = xtypes.ToPtr(upres.CertId)
	updateDomainMultiCertificatesReqBodyContent.CertName = xtypes.ToPtr(upres.CertName)
	updateDomainMultiCertificatesReqBodyContent = assign(updateDomainMultiCertificatesReqBodyContent, showDomainFullConfigResp.Configs)
	updateDomainMultiCertificatesReq := &hccdnmodel.UpdateDomainMultiCertificatesRequest{
		EnterpriseProjectId: xtypes.ToPtrOrZeroNil(d.config.EnterpriseProjectId),
		Body: &hccdnmodel.UpdateDomainMultiCertificatesRequestBody{
			Https: updateDomainMultiCertificatesReqBodyContent,
		},
	}
	updateDomainMultiCertificatesResp, err := d.sdkClient.UpdateDomainMultiCertificates(updateDomainMultiCertificatesReq)
	d.logger.Debug("sdk request 'cdn.UpdateDomainMultiCertificates'", slog.Any("request", updateDomainMultiCertificatesReq), slog.Any("response", updateDomainMultiCertificatesResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cdn.UpdateDomainMultiCertificates': %w", err)
	}

	return &core.SSLDeployResult{}, nil
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*hccdn.CdnClient, error) {
	if region == "" {
		region = "cn-north-1" // CDN 服务默认区域：华北一北京
	}

	auth, err := global.NewCredentialsBuilder().
		WithAk(accessKeyId).
		WithSk(secretAccessKey).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	hcRegion, err := hccdnregion.SafeValueOf(region)
	if err != nil {
		return nil, err
	}

	hcClient, err := hccdn.CdnClientBuilder().
		WithRegion(hcRegion).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	client := hccdn.NewCdnClient(hcClient)
	return client, nil
}

func assign(source *hccdnmodel.UpdateDomainMultiCertificatesRequestBodyContent, target *hccdnmodel.ConfigsGetBody) *hccdnmodel.UpdateDomainMultiCertificatesRequestBodyContent {
	// `UpdateDomainMultiCertificates` 中不传的字段表示使用默认值、而非保留原值，
	// 因此这里需要把原配置中的参数重新赋值回去。

	if target == nil {
		return source
	}

	if *target.OriginProtocol == "follow" {
		source.AccessOriginWay = xtypes.ToPtr(int32(1))
	} else if *target.OriginProtocol == "http" {
		source.AccessOriginWay = xtypes.ToPtr(int32(2))
	} else if *target.OriginProtocol == "https" {
		source.AccessOriginWay = xtypes.ToPtr(int32(3))
	}

	if target.ForceRedirect != nil {
		if source.ForceRedirectConfig == nil {
			source.ForceRedirectConfig = &hccdnmodel.ForceRedirect{}
		}

		if target.ForceRedirect.Status == "on" {
			source.ForceRedirectConfig.Switch = 1
			source.ForceRedirectConfig.RedirectType = target.ForceRedirect.Type
		} else {
			source.ForceRedirectConfig.Switch = 0
		}
	}

	if target.Https != nil {
		if *target.Https.Http2Status == "on" {
			source.Http2 = xtypes.ToPtr(int32(1))
		}
	}

	return source
}
