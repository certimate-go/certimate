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
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	"github.com/certimate-go/certimate/pkg/core/certmgr/providers/huaweicloud-scm/internal"
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
	sdkClient *internal.ScmClient
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
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 查询已有证书，避免重复上传
	// REF: https://support.huaweicloud.com/api-ccm/ListCertificates.html
	// REF: https://support.huaweicloud.com/api-ccm/ExportCertificate_0.html
	listCertificatesLimit := 50
	listCertificatesOffset := 0
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCertificatesReq := &hcscmmodel.ListCertificatesRequest{
			EnterpriseProjectId: lo.EmptyableToPtr(c.config.EnterpriseProjectId),
			Limit:               lo.ToPtr(int32(listCertificatesLimit)),
			Offset:              lo.ToPtr(int32(listCertificatesOffset)),
			SortDir:             lo.ToPtr("DESC"),
			SortKey:             lo.ToPtr("certExpiredTime"),
		}
		listCertificatesResp, err := c.sdkClient.ListCertificates(listCertificatesReq)
		c.logger.Debug("sdk request 'scm.ListCertificates'", slog.Any("request", listCertificatesReq), slog.Any("response", listCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'scm.ListCertificates': %w", err)
		}

		if listCertificatesResp.Certificates == nil {
			break
		}

		for _, certItem := range *listCertificatesResp.Certificates {
			// 对比证书通用名称
			if !strings.EqualFold(certX509.Subject.CommonName, certItem.Domain) {
				continue
			}

			// 对比证书有效期
			if certX509.NotAfter.Local().Format(time.DateTime) != strings.TrimSuffix(certItem.ExpireTime, ".0") {
				continue
			}

			// 对比证书内容
			exportCertificateReq := &hcscmmodel.ExportCertificateRequest{
				CertificateId: certItem.Id,
			}
			exportCertificateResp, err := c.sdkClient.ExportCertificate(exportCertificateReq)
			c.logger.Debug("sdk request 'scm.ExportCertificate'", slog.Any("request", exportCertificateReq), slog.Any("response", exportCertificateResp))
			if err != nil {
				if exportCertificateResp != nil && exportCertificateResp.HttpStatusCode == 404 {
					continue
				}
				return nil, fmt.Errorf("failed to execute sdk request 'scm.ExportCertificate': %w", err)
			} else {
				if !xcert.EqualCertificatesFromPEM(certPEM, lo.FromPtr(exportCertificateResp.Certificate)) {
					continue
				}
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			c.logger.Info("ssl certificate already exists")
			return &certmgr.UploadResult{
				CertId:   certItem.Id,
				CertName: certItem.Name,
			}, nil
		}

		if len(*listCertificatesResp.Certificates) < listCertificatesLimit {
			break
		}

		listCertificatesOffset += listCertificatesLimit
	}

	// 生成新证书名（需符合华为云命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// 上传新证书
	// REF: https://support.huaweicloud.com/api-ccm/ImportCertificate.html
	importCertificateReq := &hcscmmodel.ImportCertificateRequest{
		Body: &hcscmmodel.ImportCertificateRequestBody{
			EnterpriseProjectId: lo.EmptyableToPtr(c.config.EnterpriseProjectId),
			Name:                certName,
			Certificate:         certPEM,
			PrivateKey:          privkeyPEM,
		},
	}
	importCertificateResp, err := c.sdkClient.ImportCertificate(importCertificateReq)
	c.logger.Debug("sdk request 'scm.ImportCertificate'", slog.Any("request", importCertificateReq), slog.Any("response", importCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'scm.ImportCertificate': %w", err)
	}

	return &certmgr.UploadResult{
		CertId:   *importCertificateResp.CertificateId,
		CertName: certName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	return nil, certmgr.ErrUnsupported
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*internal.ScmClient, error) {
	if region == "" {
		region = "cn-north-4" // SCM 服务默认区域：华北北京四
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

	client := internal.NewScmClient(hcClient)
	return client, nil
}
