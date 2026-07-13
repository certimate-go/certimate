package huaweicloudvod

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	hwiam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	hwiammodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	hwiamregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"
	hwvodmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vod/v1/model"
	hwvodregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vod/v1/region"
	"github.com/samber/lo"

	hwvod "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vod/v1"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/huaweicloud-scm"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// 华为云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 华为云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 华为云企业项目 ID。
	EnterpriseProjectId string `json:"enterpriseProjectId,omitempty"`
	// 华为云区域。
	Region string `json:"region"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 点播加速域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *hwvod.VodClient
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(
		config.AccessKeyId,
		config.SecretAccessKey,
		config.Region,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		AccessKeyId:         config.AccessKeyId,
		SecretAccessKey:     config.SecretAccessKey,
		EnterpriseProjectId: config.EnterpriseProjectId,
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
	var domains []string
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if d.config.Domain == "" {
				return nil, fmt.Errorf("config `domain` is required")
			}

			domains = []string{d.config.Domain}
		}

	default:
		return nil, fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 遍历更新域名证书
	if len(domains) == 0 {
		d.logger.Info("no vod domains to deploy")
	} else {
		d.logger.Info("found vod domains to deploy", slog.Any("domains", domains))
		var errs []error

		for _, domain := range domains {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				if err := d.updateDomainsCertificate(ctx, domain, upres.CertId); err != nil {
					errs = append(errs, err)
				}
			}
		}

		if len(errs) > 0 {
			return nil, errors.Join(errs...)
		}
	}

	return &DeployResult{}, nil
}

func (d *Deployer) updateDomainsCertificate(ctx context.Context, domain string, cloudCertId string) error {
	// 查询 HTTPS 配置
	// REF: https://support.huaweicloud.com/api-vod/ShowHttpsConfig.html
	showHttpsConfigReq := &hwvodmodel.ShowHttpsConfigRequest{
		Domain: domain,
	}
	showHttpsConfigResp, err := d.sdkClient.ShowHttpsConfig(showHttpsConfigReq)
	d.logger.Debug("sdk request 'vod.ShowHttpsConfig'", slog.Any("request", showHttpsConfigReq), slog.Any("response", showHttpsConfigResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'vod.ShowHttpsConfig': %w", err)
	} else if lo.FromPtr(showHttpsConfigResp.CertId) == cloudCertId {
		// 已部署过，直接返回
		return nil
	}

	// 配置 HTTPS
	// REF: https://support.huaweicloud.com/api-vod/UpdateHttpsConfig.html
	updateHttpsConfigReq := &hwvodmodel.UpdateHttpsConfigRequest{
		Body: &hwvodmodel.ConfigCdnHttpsReq{
			Domain:             domain,
			Source:             lo.ToPtr("scm"),
			CertId:             lo.ToPtr(cloudCertId),
			HttpsStatus:        lo.ToPtr(int32(1)),
			Http2:              showHttpsConfigResp.Http2,
			ForceRedirectHttps: showHttpsConfigResp.ForceRedirectHttps,
		},
	}
	updateHttpsConfigResp, err := d.sdkClient.UpdateHttpsConfig(updateHttpsConfigReq)
	d.logger.Debug("sdk request 'vod.UpdateHttpsConfig'", slog.Any("request", updateHttpsConfigReq), slog.Any("response", updateHttpsConfigResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'vod.UpdateHttpsConfig': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*hwvod.VodClient, error) {
	projectId, err := getSDKProjectId(accessKeyId, secretAccessKey, region)
	if err != nil {
		return nil, err
	}

	auth, err := basic.NewCredentialsBuilder().
		WithAk(accessKeyId).
		WithSk(secretAccessKey).
		WithProjectId(projectId).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	hcRegion, err := hwvodregion.SafeValueOf(region)
	if err != nil {
		return nil, err
	}

	hcClient, err := hwvod.VodClientBuilder().
		WithRegion(hcRegion).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	client := hwvod.NewVodClient(hcClient)
	return client, nil
}

func getSDKProjectId(accessKeyId, secretAccessKey, region string) (string, error) {
	auth, err := global.NewCredentialsBuilder().
		WithAk(accessKeyId).
		WithSk(secretAccessKey).
		SafeBuild()
	if err != nil {
		return "", err
	}

	hcRegion, err := hwiamregion.SafeValueOf(region)
	if err != nil {
		return "", err
	}

	hcClient, err := hwiam.IamClientBuilder().
		WithRegion(hcRegion).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return "", err
	}

	client := hwiam.NewIamClient(hcClient)

	request := &hwiammodel.KeystoneListProjectsRequest{
		Name: &region,
	}
	response, err := client.KeystoneListProjects(request)
	if err != nil {
		return "", err
	} else if response.Projects == nil || len(*response.Projects) == 0 {
		return "", fmt.Errorf("huaweicloud: no project found")
	}

	return (*response.Projects)[0].Id, nil
}
