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

	"github.com/certimate-go/certimate/pkg/core"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xtypes "github.com/certimate-go/certimate/pkg/utils/types"
)

type SSLManagerProviderConfig struct {
	// 华为云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 华为云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 华为云企业项目 ID。
	EnterpriseProjectId string `json:"enterpriseProjectId,omitempty"`
	// 华为云区域。
	Region string `json:"region"`
}

type SSLManagerProvider struct {
	config    *SSLManagerProviderConfig
	logger    *slog.Logger
	sdkClient *hcelb.ElbClient
}

var _ core.SSLManager = (*SSLManagerProvider)(nil)

func NewSSLManagerProvider(config *SSLManagerProviderConfig) (*SSLManagerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl manager provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create sdk client: %w", err)
	}

	return &SSLManagerProvider{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (m *SSLManagerProvider) SetLogger(logger *slog.Logger) {
	if logger == nil {
		m.logger = slog.New(slog.DiscardHandler)
	} else {
		m.logger = logger
	}
}

func (m *SSLManagerProvider) Upload(ctx context.Context, certPEM string, privkeyPEM string) (*core.SSLManageUploadResult, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 遍历查询已有证书，避免重复上传
	// REF: https://support.huaweicloud.com/api-elb/ListCertificates.html
	listCertificatesLimit := int32(2000)
	var listCertificatesMarker *string = nil
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCertificatesReq := &hcelbmodel.ListCertificatesRequest{
			Limit:  xtypes.ToPtr(listCertificatesLimit),
			Marker: listCertificatesMarker,
			Type:   &[]string{"server"},
		}
		listCertificatesResp, err := m.sdkClient.ListCertificates(listCertificatesReq)
		m.logger.Debug("sdk request 'elb.ListCertificates'", slog.Any("request", listCertificatesReq), slog.Any("response", listCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'elb.ListCertificates': %w", err)
		}

		if listCertificatesResp.Certificates != nil {
			for _, certInfo := range *listCertificatesResp.Certificates {
				var isSameCert bool
				if certInfo.Certificate == certPEM {
					isSameCert = true
				} else {
					oldCertX509, err := xcert.ParseCertificateFromPEM(certInfo.Certificate)
					if err != nil {
						continue
					}

					isSameCert = xcert.EqualCertificate(certX509, oldCertX509)
				}

				// 如果已存在相同证书，直接返回
				if isSameCert {
					m.logger.Info("ssl certificate already exists")
					return &core.SSLManageUploadResult{
						CertId:   certInfo.Id,
						CertName: certInfo.Name,
					}, nil
				}
			}
		}

		if listCertificatesResp.Certificates == nil || len(*listCertificatesResp.Certificates) < int(listCertificatesLimit) {
			break
		} else {
			listCertificatesMarker = listCertificatesResp.PageInfo.NextMarker
		}
	}

	// 获取项目 ID
	// REF: https://support.huaweicloud.com/api-iam/iam_06_0001.html
	projectId, err := getSdkProjectId(m.config.AccessKeyId, m.config.SecretAccessKey, m.config.Region)
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
				EnterpriseProjectId: xtypes.ToPtrOrZeroNil(m.config.EnterpriseProjectId),
				ProjectId:           xtypes.ToPtr(projectId),
				Name:                xtypes.ToPtr(certName),
				Certificate:         xtypes.ToPtr(certPEM),
				PrivateKey:          xtypes.ToPtr(privkeyPEM),
			},
		},
	}
	createCertificateResp, err := m.sdkClient.CreateCertificate(createCertificateReq)
	m.logger.Debug("sdk request 'elb.CreateCertificate'", slog.Any("request", createCertificateReq), slog.Any("response", createCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'elb.CreateCertificate': %w", err)
	}

	return &core.SSLManageUploadResult{
		CertId:   createCertificateResp.Certificate.Id,
		CertName: createCertificateResp.Certificate.Name,
	}, nil
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*hcelb.ElbClient, error) {
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

	client := hcelb.NewElbClient(hcClient)
	return client, nil
}

func getSdkProjectId(accessKeyId, secretAccessKey, region string) (string, error) {
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
		return "", errors.New("no project found")
	}

	return (*response.Projects)[0].Id, nil
}
