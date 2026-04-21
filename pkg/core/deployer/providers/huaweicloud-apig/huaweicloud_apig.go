package huaweicloudapig

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	hcapig "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/apig/v2"
	hcapigmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/apig/v2/model"
	hcapigregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/apig/v2/region"
	hciam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	hciamModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	hciamregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/huaweicloud-apig/internal"
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
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *internal.ApigClient
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(
		config.AccessKeyId,
		config.SecretAccessKey,
		config.Region,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return &Deployer{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (d *Deployer) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM, issuerCertificatePEM string) (*deployer.DeployResult, error) {
	// 根据部署资源类型决定部署方式
	switch d.config.ResourceType {
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

	// 查询证书详情
	// REF: https://support.huaweicloud.com/api-apig/ShowDetailsOfCertificateV2.html
	showDetailsOfCertificateV2Req := &hcapigmodel.ShowDetailsOfCertificateV2Request{
		CertificateId: d.config.CertificateId,
	}
	showDetailsOfCertificateV2Resp, err := d.sdkClient.ShowDetailsOfCertificateV2(showDetailsOfCertificateV2Req)
	d.logger.Debug("sdk request 'apig.ShowDetailsOfCertificateV2'", slog.Any("request", showDetailsOfCertificateV2Req), slog.Any("response", showDetailsOfCertificateV2Resp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'apig.ShowDetailsOfCertificateV2': %w", err)
	}

	// 修改 SSL 证书
	// REF: https://support.huaweicloud.com/api-apig/UpdateCertificateV2.html
	updateCertificateV2Req := &hcapigmodel.UpdateCertificateV2Request{
		CertificateId: d.config.CertificateId,
		Body: &hcapigmodel.CertificateForm{
			Name:        fmt.Sprintf("certimate_%d", time.Now().UnixMilli()),
			CertContent: certPEM,
			PrivateKey:  privkeyPEM,
			Type: lo.
				If(showDetailsOfCertificateV2Resp.Type.Value() == hcapigmodel.GetCertificateFormTypeEnum().INSTANCE.Value(), lo.ToPtr(hcapigmodel.GetCertificateFormTypeEnum().INSTANCE)).
				Else(lo.ToPtr(hcapigmodel.GetCertificateFormTypeEnum().GLOBAL)),
			InstanceId: showDetailsOfCertificateV2Resp.InstanceId,
		},
	}
	updateCertificateV2Resp, err := d.sdkClient.UpdateCertificateV2(updateCertificateV2Req)
	d.logger.Debug("sdk request 'apig.UpdateCertificateV2'", slog.Any("request", updateCertificateV2Req), slog.Any("response", updateCertificateV2Resp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'apig.UpdateCertificateV2': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*internal.ApigClient, error) {
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

	hcRegion, err := hcapigregion.SafeValueOf(region)
	if err != nil {
		return nil, err
	}

	hcClient, err := hcapig.ApigClientBuilder().
		WithRegion(hcRegion).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	client := internal.NewApigClient(hcClient)
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
