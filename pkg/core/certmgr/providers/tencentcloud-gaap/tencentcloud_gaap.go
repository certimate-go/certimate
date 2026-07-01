package tencentcloudgaap

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcerrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	tcgaap "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/gaap/v20180529"

	"github.com/certimate-go/certimate/pkg/core"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type (
	Provider      = core.Certmgr
	UploadResult  = core.CertmgrUploadResult
	ReplaceResult = core.CertmgrReplaceResult
)

type CertmgrConfig struct {
	// 腾讯云 SecretId。
	SecretId string `json:"secretId"`
	// 腾讯云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 腾讯云项目 ID。
	ProjectId int64 `json:"projectId,omitempty"`
	// 腾讯云接口端点。
	Endpoint string `json:"endpoint,omitempty"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *tcgaap.Client
}

var _ Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.SecretId, config.SecretKey, config.Endpoint)
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

func (c *Certmgr) Upload(ctx context.Context, certPEM, privkeyPEM string) (*UploadResult, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 查询服务器证书列表，避免重复上传
	// REF: https://cloud.tencent.com/document/api/1364/98588
	// REF: https://cloud.tencent.com/document/api/608/36978
	describeCertificatesOffset := 0
	describeCertificatesLimit := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		describeCertificatesReq := tcgaap.NewDescribeCertificatesRequest()
		describeCertificatesReq.CertificateType = common.Int64Ptr(2)
		describeCertificatesReq.Offset = common.Uint64Ptr(uint64(describeCertificatesOffset))
		describeCertificatesReq.Limit = common.Uint64Ptr(uint64(describeCertificatesLimit))
		describeCertificatesResp, err := c.sdkClient.DescribeCertificatesWithContext(ctx, describeCertificatesReq)
		c.logger.Debug("sdk request 'gaap.DescribeCertificates'", slog.Any("request", describeCertificatesReq), slog.Any("response", describeCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'gaap.DescribeCertificates': %w", err)
		}

		for _, certItem := range describeCertificatesResp.Response.CertificateSet {
			// 对比证书通用名称
			if !strings.EqualFold(certX509.Subject.CommonName, lo.FromPtr(certItem.SubjectCN)) {
				continue
			}

			// 对比证书有效期
			if certX509.NotBefore.Unix() != int64(lo.FromPtr(certItem.BeginTime)) ||
				certX509.NotAfter.Unix() != int64(lo.FromPtr(certItem.EndTime)) {
				continue
			}

			// 对比证书内容
			describeCertificateDetailReq := tcgaap.NewDescribeCertificateDetailRequest()
			describeCertificateDetailReq.CertificateId = certItem.CertificateId
			describeCertificateDetailResp, err := c.sdkClient.DescribeCertificateDetailWithContext(ctx, describeCertificateDetailReq)
			c.logger.Debug("sdk request 'gaap.DescribeCertificateDetail'", slog.Any("request", describeCertificateDetailReq), slog.Any("response", describeCertificateDetailResp))
			if err != nil {
				if sdkErr, ok := err.(*tcerrors.TencentCloudSDKError); ok {
					if sdkErrCode := sdkErr.Code; sdkErrCode == "ResourceNotFound" {
						continue
					}
				}

				return nil, fmt.Errorf("failed to execute sdk request 'gaap.DescribeCertificateDetail': %w", err)
			} else {
				if !xcert.EqualCertificatesFromPEM(certPEM, lo.FromPtr(describeCertificateDetailResp.Response.CertificateDetail.CertificateContent)) {
					continue
				}
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			c.logger.Info("ssl certificate already exists")
			return &UploadResult{
				CertId:   lo.FromPtr(certItem.CertificateId),
				CertName: lo.FromPtr(certItem.CertificateAlias),
			}, nil
		}

		if len(describeCertificatesResp.Response.CertificateSet) < describeCertificatesLimit {
			break
		}

		describeCertificatesOffset += describeCertificatesLimit
	}

	// 生成新证书名（需符合腾讯云命名规则）
	certName := fmt.Sprintf("certimate_%d", time.Now().UnixMilli())

	// 创建云原生网关证书
	// REF: https://cloud.tencent.com/document/api/608/36980
	createCertificateReq := tcgaap.NewCreateCertificateRequest()
	createCertificateReq.CertificateType = common.Int64Ptr(2)
	createCertificateReq.CertificateAlias = common.StringPtr(certName)
	createCertificateReq.CertificateContent = common.StringPtr(certPEM)
	createCertificateReq.CertificateKey = common.StringPtr(privkeyPEM)
	createCertificateResp, err := c.sdkClient.CreateCertificateWithContext(ctx, createCertificateReq)
	c.logger.Debug("sdk request 'gaap.CreateCertificate'", slog.Any("request", createCertificateReq), slog.Any("response", createCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'gaap.CreateCertificate': %w", err)
	}

	return &UploadResult{
		CertId:   lo.FromPtr(createCertificateResp.Response.CertificateId),
		CertName: certName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*ReplaceResult, error) {
	return nil, core.ErrUnsupported
}

func createSDKClient(secretId, secretKey, endpoint string) (*tcgaap.Client, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := tcgaap.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
