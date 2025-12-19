package huaweicloudelb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	hcelb "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3"
	hcelbmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3/model"
	hcelbregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3/region"
	hciam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	hciammodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	hciamregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	"github.com/certimate-go/certimate/pkg/core/certmgr/providers/huaweicloud-elb/internal"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type CertmgrConfig struct {
	// 华为云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 华为云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 华为云企业项目 ID。
	EnterpriseProjectId string `json:"enterpriseProjectId,omitempty"`
	// 华为云区域。
	Region string `json:"region"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *internal.ElbClient
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, errors.New("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return &Certmgr{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (c *Certmgr) SetLogger(logger *slog.Logger) {
	if logger == nil {
		c.logger = slog.New(slog.DiscardHandler)
	} else {
		c.logger = logger
	}
}

func (c *Certmgr) Upload(ctx context.Context, certPEM, privkeyPEM string) (*certmgr.UploadResult, error) {
	// 查询已有证书，避免重复上传
	// REF: https://support.huaweicloud.com/api-elb/ListCertificates.html
	listCertificatesMarker := (*string)(nil)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCertificatesReq := &hcelbmodel.ListCertificatesRequest{
			Marker: listCertificatesMarker,
			Limit:  lo.ToPtr(int32(2000)),
			Type:   lo.ToPtr([]string{"server"}),
		}
		listCertificatesResp, err := c.sdkClient.ListCertificates(listCertificatesReq)
		c.logger.Debug("sdk request 'elb.ListCertificates'", slog.Any("request", listCertificatesReq), slog.Any("response", listCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'elb.ListCertificates': %w", err)
		}

		if listCertificatesResp.Certificates == nil {
			break
		}

		for _, certItem := range *listCertificatesResp.Certificates {
			// 如果已存在相同证书，直接返回
			if xcert.EqualCertificatesFromPEM(certPEM, certItem.Certificate) {
				c.logger.Info("ssl certificate already exists")
				return &certmgr.UploadResult{
					CertId:   certItem.Id,
					CertName: certItem.Name,
				}, nil
			}
		}

		if len(*listCertificatesResp.Certificates) == 0 || listCertificatesResp.PageInfo.NextMarker == nil {
			break
		}

		listCertificatesMarker = listCertificatesResp.PageInfo.NextMarker
	}

	// 获取项目 ID
	// REF: https://support.huaweicloud.com/api-iam/iam_06_0001.html
	projectId, err := getSDKProjectId(c.config.AccessKeyId, c.config.SecretAccessKey, c.config.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to get SDK project id: %w", err)
	}

	// 生成新证书名（需符合华为云命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// 创建新证书
	// REF: https://support.huaweicloud.com/api-elb/CreateCertificate.html
	createCertificateReq := &hcelbmodel.CreateCertificateRequest{
		Body: &hcelbmodel.CreateCertificateRequestBody{
			Certificate: &hcelbmodel.CreateCertificateOption{
				EnterpriseProjectId: lo.EmptyableToPtr(c.config.EnterpriseProjectId),
				ProjectId:           lo.ToPtr(projectId),
				Name:                lo.ToPtr(certName),
				Certificate:         lo.ToPtr(certPEM),
				PrivateKey:          lo.ToPtr(privkeyPEM),
			},
		},
	}
	createCertificateResp, err := c.sdkClient.CreateCertificate(createCertificateReq)
	c.logger.Debug("sdk request 'elb.CreateCertificate'", slog.Any("request", createCertificateReq), slog.Any("response", createCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'elb.CreateCertificate': %w", err)
	}

	return &certmgr.UploadResult{
		CertId:   createCertificateResp.Certificate.Id,
		CertName: createCertificateResp.Certificate.Name,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	// 更新证书
	// REF: https://support.huaweicloud.com/api-elb/UpdateCertificate.html
	updateCertificateReq := &hcelbmodel.UpdateCertificateRequest{
		CertificateId: certIdOrName,
		Body: &hcelbmodel.UpdateCertificateRequestBody{
			Certificate: &hcelbmodel.UpdateCertificateOption{
				Certificate: lo.ToPtr(certPEM),
				PrivateKey:  lo.ToPtr(privkeyPEM),
			},
		},
	}
	updateCertificateResp, err := c.sdkClient.UpdateCertificate(updateCertificateReq)
	c.logger.Debug("sdk request 'elb.UpdateCertificate'", slog.Any("request", updateCertificateReq), slog.Any("response", updateCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'elb.UpdateCertificate': %w", err)
	}

	return &certmgr.OperateResult{}, nil
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*internal.ElbClient, error) {
	if region == "" {
		region = "cn-north-4" // ELB 服务默认区域：华北四北京
	}

	auth, err := basic.NewCredentialsBuilder().
		WithAk(accessKeyId).
		WithSk(secretAccessKey).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	hcRegion, err := hcelbregion.SafeValueOf(region)
	if err != nil {
		return nil, err
	}

	hcClient, err := hcelb.ElbClientBuilder().
		WithRegion(hcRegion).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	client := internal.NewElbClient(hcClient)
	return client, nil
}

func getSDKProjectId(accessKeyId, secretAccessKey, region string) (string, error) {
	if region == "" {
		region = "cn-north-4" // IAM 服务默认区域：华北四北京
	}

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

	request := &hciammodel.KeystoneListProjectsRequest{
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
