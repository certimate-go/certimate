package huaweicloudscm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	hcscm "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/scm/v3"
	hcscmmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/scm/v3/model"
	hcscmregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/scm/v3/region"

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
	sdkClient *hcscm.ScmClient
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
	// REF: https://support.huaweicloud.com/api-ccm/ListCertificates.html
	// REF: https://support.huaweicloud.com/api-ccm/ExportCertificate_0.html
	listCertificatesLimit := int32(50)
	listCertificatesOffset := int32(0)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCertificatesReq := &hcscmmodel.ListCertificatesRequest{
			EnterpriseProjectId: xtypes.ToPtrOrZeroNil(m.config.EnterpriseProjectId),
			Limit:               xtypes.ToPtr(listCertificatesLimit),
			Offset:              xtypes.ToPtr(listCertificatesOffset),
			SortDir:             xtypes.ToPtr("DESC"),
			SortKey:             xtypes.ToPtr("certExpiredTime"),
		}
		listCertificatesResp, err := m.sdkClient.ListCertificates(listCertificatesReq)
		m.logger.Debug("sdk request 'scm.ListCertificates'", slog.Any("request", listCertificatesReq), slog.Any("response", listCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'scm.ListCertificates': %w", err)
		}

		if listCertificatesResp.Certificates != nil {
			for _, certDetail := range *listCertificatesResp.Certificates {
				// 先对比证书通用名称
				if !strings.EqualFold(certX509.Subject.CommonName, certDetail.Domain) {
					continue
				}

				// 再对比证书有效期
				if certX509.NotAfter.Local().Format(time.DateTime) != strings.TrimSuffix(certDetail.ExpireTime, ".0") {
					continue
				}

				// 最后对比证书内容
				exportCertificateReq := &hcscmmodel.ExportCertificateRequest{
					CertificateId: certDetail.Id,
				}
				exportCertificateResp, err := m.sdkClient.ExportCertificate(exportCertificateReq)
				m.logger.Debug("sdk request 'scm.ExportCertificate'", slog.Any("request", exportCertificateReq), slog.Any("response", exportCertificateResp))
				if err != nil {
					if exportCertificateResp != nil && exportCertificateResp.HttpStatusCode == 404 {
						continue
					}
					return nil, fmt.Errorf("failed to execute sdk request 'scm.ExportCertificate': %w", err)
				} else {
					var isSameCert bool
					if *exportCertificateResp.Certificate == certPEM {
						isSameCert = true
					} else {
						oldCertX509, err := xcert.ParseCertificateFromPEM(*exportCertificateResp.Certificate)
						if err != nil {
							continue
						}

						isSameCert = xcert.EqualCertificate(certX509, oldCertX509)
					}

					// 如果已存在相同证书，直接返回
					if isSameCert {
						m.logger.Info("ssl certificate already exists")
						return &core.SSLManageUploadResult{
							CertId:   certDetail.Id,
							CertName: certDetail.Name,
						}, nil
					}
				}
			}
		}

		if listCertificatesResp.Certificates == nil || len(*listCertificatesResp.Certificates) < int(listCertificatesLimit) {
			break
		} else {
			listCertificatesOffset += listCertificatesLimit
		}
	}

	// 生成新证书名（需符合华为云命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// 上传新证书
	// REF: https://support.huaweicloud.com/api-ccm/ImportCertificate.html
	importCertificateReq := &hcscmmodel.ImportCertificateRequest{
		Body: &hcscmmodel.ImportCertificateRequestBody{
			EnterpriseProjectId: xtypes.ToPtrOrZeroNil(m.config.EnterpriseProjectId),
			Name:                certName,
			Certificate:         certPEM,
			PrivateKey:          privkeyPEM,
		},
	}
	importCertificateResp, err := m.sdkClient.ImportCertificate(importCertificateReq)
	m.logger.Debug("sdk request 'scm.ImportCertificate'", slog.Any("request", importCertificateReq), slog.Any("response", importCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'scm.ImportCertificate': %w", err)
	}

	return &core.SSLManageUploadResult{
		CertId:   *importCertificateResp.CertificateId,
		CertName: certName,
	}, nil
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*hcscm.ScmClient, error) {
	if region == "" {
		region = "cn-north-4" // SCM 服务默认区域：华北四北京
	}

	auth, err := basic.NewCredentialsBuilder().
		WithAk(accessKeyId).
		WithSk(secretAccessKey).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	hcRegion, err := hcscmregion.SafeValueOf(region)
	if err != nil {
		return nil, err
	}

	hcClient, err := hcscm.ScmClientBuilder().
		WithRegion(hcRegion).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	client := hcscm.NewScmClient(hcClient)
	return client, nil
}
