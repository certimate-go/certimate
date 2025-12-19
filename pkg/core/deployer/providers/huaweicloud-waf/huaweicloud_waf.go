package huaweicloudwaf

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	hciam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	hciamModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	hciamregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"
	hcwaf "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/waf/v1"
	hcwafmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/waf/v1/model"
	hcwafregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/waf/v1/region"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/huaweicloud-waf"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/huaweicloud-waf/internal"
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
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// 证书 ID。
	// 部署资源类型为 [RESOURCE_TYPE_CERTIFICATE] 时必填。
	CertificateId string `json:"certificateId,omitempty"`
	// 防护域名（支持泛域名）。
	// 部署资源类型为 [RESOURCE_TYPE_CLOUDSERVER]、[RESOURCE_TYPE_PREMIUMHOST] 时必填。
	Domain string `json:"domain,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *internal.WafClient
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		AccessKeyId:         config.AccessKeyId,
		SecretAccessKey:     config.SecretAccessKey,
		EnterpriseProjectId: config.EnterpriseProjectId,
		Region:              config.Region,
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
	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 根据部署资源类型决定部署方式
	switch d.config.ResourceType {
	case RESOURCE_TYPE_CLOUDSERVER:
		if err := d.deployToCloudServer(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	case RESOURCE_TYPE_PREMIUMHOST:
		if err := d.deployToPremiumHost(ctx, certPEM, privkeyPEM); err != nil {
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

func (d *Deployer) deployToCertificate(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.CertificateId == "" {
		return errors.New("config `certificateId` is required")
	}

	// 查询证书
	// REF: https://support.huaweicloud.com/api-waf/ShowCertificate.html
	showCertificateReq := &hcwafmodel.ShowCertificateRequest{
		EnterpriseProjectId: lo.EmptyableToPtr(d.config.EnterpriseProjectId),
		CertificateId:       d.config.CertificateId,
	}
	showCertificateResp, err := d.sdkClient.ShowCertificate(showCertificateReq)
	d.logger.Debug("sdk request 'waf.ShowCertificate'", slog.Any("request", showCertificateReq), slog.Any("response", showCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'waf.ShowCertificate': %w", err)
	}

	// 更新证书
	// REF: https://support.huaweicloud.com/api-waf/UpdateCertificate.html
	updateCertificateReq := &hcwafmodel.UpdateCertificateRequest{
		EnterpriseProjectId: lo.EmptyableToPtr(d.config.EnterpriseProjectId),
		CertificateId:       d.config.CertificateId,
		Body: &hcwafmodel.UpdateCertificateRequestBody{
			Name:    *showCertificateResp.Name,
			Content: lo.ToPtr(certPEM),
			Key:     lo.ToPtr(privkeyPEM),
		},
	}
	updateCertificateResp, err := d.sdkClient.UpdateCertificate(updateCertificateReq)
	d.logger.Debug("sdk request 'waf.UpdateCertificate'", slog.Any("request", updateCertificateReq), slog.Any("response", updateCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'waf.UpdateCertificate': %w", err)
	}

	return nil
}

func (d *Deployer) deployToCloudServer(ctx context.Context, certPEM, privkeyPEM string) error {
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

	// 查询云模式防护域名列表，获取防护域名 ID
	// REF: https://support.huaweicloud.com/api-waf/ListHost.html
	hostId := ""
	listHostPage := 1
	listHostPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		listHostReq := &hcwafmodel.ListHostRequest{
			EnterpriseProjectId: lo.EmptyableToPtr(d.config.EnterpriseProjectId),
			Hostname:            lo.ToPtr(strings.TrimPrefix(d.config.Domain, "*")),
			Page:                lo.ToPtr(int32(listHostPage)),
			Pagesize:            lo.ToPtr(int32(listHostPageSize)),
		}
		listHostResp, err := d.sdkClient.ListHost(listHostReq)
		d.logger.Debug("sdk request 'waf.ListHost'", slog.Any("request", listHostReq), slog.Any("response", listHostResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'waf.ListHost': %w", err)
		}

		if listHostResp.Items == nil {
			break
		}

		for _, hostItem := range *listHostResp.Items {
			if strings.TrimPrefix(d.config.Domain, "*") == *hostItem.Hostname {
				hostId = *hostItem.Id
				break
			}
		}

		if len(*listHostResp.Items) < listHostPageSize {
			break
		}

		listHostPage++
	}
	if hostId == "" {
		return fmt.Errorf("could not find cloudserver host '%s'", d.config.Domain)
	}

	// 更新云模式防护域名的配置
	// REF: https://support.huaweicloud.com/api-waf/UpdateHost.html
	updateHostReq := &hcwafmodel.UpdateHostRequest{
		EnterpriseProjectId: lo.EmptyableToPtr(d.config.EnterpriseProjectId),
		InstanceId:          hostId,
		Body: &hcwafmodel.UpdateHostRequestBody{
			Certificateid:   lo.ToPtr(upres.CertId),
			Certificatename: lo.ToPtr(upres.CertName),
		},
	}
	updateHostResp, err := d.sdkClient.UpdateHost(updateHostReq)
	d.logger.Debug("sdk request 'waf.UpdateHost'", slog.Any("request", updateHostReq), slog.Any("response", updateHostResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'waf.UpdateHost': %w", err)
	}

	return nil
}

func (d *Deployer) deployToPremiumHost(ctx context.Context, certPEM, privkeyPEM string) error {
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

	// 查询独享模式域名列表，获取防护域名 ID
	// REF: https://support.huaweicloud.com/api-waf/ListPremiumHost.html
	hostId := ""
	listPremiumHostPage := 1
	listPremiumHostPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		listPremiumHostReq := &hcwafmodel.ListPremiumHostRequest{
			EnterpriseProjectId: lo.EmptyableToPtr(d.config.EnterpriseProjectId),
			Hostname:            lo.ToPtr(strings.TrimPrefix(d.config.Domain, "*")),
			Page:                lo.ToPtr(fmt.Sprintf("%d", listPremiumHostPage)),
			Pagesize:            lo.ToPtr(fmt.Sprintf("%d", listPremiumHostPageSize)),
		}
		listPremiumHostResp, err := d.sdkClient.ListPremiumHost(listPremiumHostReq)
		d.logger.Debug("sdk request 'waf.ListPremiumHost'", slog.Any("request", listPremiumHostReq), slog.Any("response", listPremiumHostResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'waf.ListPremiumHost': %w", err)
		}

		if listPremiumHostResp.Items == nil {
			break
		}

		for _, hostItem := range *listPremiumHostResp.Items {
			if strings.TrimPrefix(d.config.Domain, "*") == *hostItem.Hostname {
				hostId = *hostItem.Id
				break
			}
		}

		if len(*listPremiumHostResp.Items) < listPremiumHostPageSize {
			break
		}

		listPremiumHostPage++
	}
	if hostId == "" {
		return fmt.Errorf("could not find premium host '%s'", d.config.Domain)
	}

	// 修改独享模式域名配置
	// REF: https://support.huaweicloud.com/api-waf/UpdatePremiumHost.html
	updatePremiumHostReq := &hcwafmodel.UpdatePremiumHostRequest{
		EnterpriseProjectId: lo.EmptyableToPtr(d.config.EnterpriseProjectId),
		HostId:              hostId,
		Body: &hcwafmodel.UpdatePremiumHostRequestBody{
			Certificateid:   lo.ToPtr(upres.CertId),
			Certificatename: lo.ToPtr(upres.CertName),
		},
	}
	updatePremiumHostResp, err := d.sdkClient.UpdatePremiumHost(updatePremiumHostReq)
	d.logger.Debug("sdk request 'waf.UpdatePremiumHost'", slog.Any("request", updatePremiumHostReq), slog.Any("response", updatePremiumHostResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'waf.UpdatePremiumHost': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*internal.WafClient, error) {
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

	hcRegion, err := hcwafregion.SafeValueOf(region)
	if err != nil {
		return nil, err
	}

	hcClient, err := hcwaf.WafClientBuilder().
		WithRegion(hcRegion).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	client := internal.NewWafClient(hcClient)
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

	hcRegion, err := hciamregion.SafeValueOf(region)
	if err != nil {
		return "", err
	}

	hcClient, err := hciam.IamClientBuilder().
		WithRegion(hcRegion).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return "", err
	}

	client := hciam.NewIamClient(hcClient)

	request := &hciamModel.KeystoneListProjectsRequest{
		Name: &region,
	}
	response, err := client.KeystoneListProjects(request)
	if err != nil {
		return "", err
	} else if response.Projects == nil || len(*response.Projects) == 0 {
		return "", errors.New("huaweicloud: no project found")
	}

	return (*response.Projects)[0].Id, nil
}
