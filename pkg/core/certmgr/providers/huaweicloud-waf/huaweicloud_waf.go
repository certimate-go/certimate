package huaweicloudwaf

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	hciam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	hciammodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	hciamregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"
	hcwaf "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/waf/v1"
	hcwafmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/waf/v1/model"
	hcwafregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/waf/v1/region"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	"github.com/certimate-go/certimate/pkg/core/certmgr/providers/huaweicloud-waf/internal"
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
	sdkClient *internal.WafClient
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
	// REF: https://support.huaweicloud.com/api-waf/ListCertificates.html
	// REF: https://support.huaweicloud.com/api-waf/ShowCertificate.html
	listCertificatesPage := 1
	listCertificatesPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCertificatesReq := &hcwafmodel.ListCertificatesRequest{
			EnterpriseProjectId: lo.EmptyableToPtr(c.config.EnterpriseProjectId),
			Page:                lo.ToPtr(int32(listCertificatesPage)),
			Pagesize:            lo.ToPtr(int32(listCertificatesPageSize)),
		}
		listCertificatesResp, err := c.sdkClient.ListCertificates(listCertificatesReq)
		c.logger.Debug("sdk request 'waf.ShowCertificate'", slog.Any("request", listCertificatesReq), slog.Any("response", listCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'waf.ListCertificates': %w", err)
		}

		if listCertificatesResp.Items == nil {
			break
		}

		for _, certItem := range *listCertificatesResp.Items {
			showCertificateReq := &hcwafmodel.ShowCertificateRequest{
				EnterpriseProjectId: lo.EmptyableToPtr(c.config.EnterpriseProjectId),
				CertificateId:       certItem.Id,
			}
			showCertificateResp, err := c.sdkClient.ShowCertificate(showCertificateReq)
			c.logger.Debug("sdk request 'waf.ShowCertificate'", slog.Any("request", showCertificateReq), slog.Any("response", showCertificateResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request 'waf.ShowCertificate': %w", err)
			}

			// 如果已存在相同证书，直接返回
			if xcert.EqualCertificatesFromPEM(certPEM, lo.FromPtr(showCertificateResp.Content)) {
				c.logger.Info("ssl certificate already exists")
				return &certmgr.UploadResult{
					CertId:   certItem.Id,
					CertName: certItem.Name,
				}, nil
			}
		}

		if len(*listCertificatesResp.Items) < listCertificatesPageSize {
			break
		}

		listCertificatesPage++
	}

	// 生成新证书名（需符合华为云命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// 创建证书
	// REF: https://support.huaweicloud.com/api-waf/CreateCertificate.html
	createCertificateReq := &hcwafmodel.CreateCertificateRequest{
		EnterpriseProjectId: lo.EmptyableToPtr(c.config.EnterpriseProjectId),
		Body: &hcwafmodel.CreateCertificateRequestBody{
			Name:    certName,
			Content: certPEM,
			Key:     privkeyPEM,
		},
	}
	createCertificateResp, err := c.sdkClient.CreateCertificate(createCertificateReq)
	c.logger.Debug("sdk request 'waf.CreateCertificate'", slog.Any("request", createCertificateReq), slog.Any("response", createCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'waf.CreateCertificate': %w", err)
	}

	return &certmgr.UploadResult{
		CertId:   *createCertificateResp.Id,
		CertName: certName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	return nil, certmgr.ErrUnsupported
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
