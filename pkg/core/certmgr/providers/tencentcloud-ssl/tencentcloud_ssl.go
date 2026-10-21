package tencentcloudssl

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	tcssl "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"

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
	sdkClient *tcssl.Client
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
	// 上传新证书
	// REF: https://cloud.tencent.com/document/api/400/41665
	uploadCertificateReq := tcssl.NewUploadCertificateRequest()
	uploadCertificateReq.ProjectId = lo.EmptyableToPtr(uint64(c.config.ProjectId))
	uploadCertificateReq.CertificatePublicKey = common.StringPtr(certPEM)
	uploadCertificateReq.CertificatePrivateKey = common.StringPtr(privkeyPEM)
	uploadCertificateReq.Repeatable = common.BoolPtr(false)
	uploadCertificateResp, err := c.sdkClient.UploadCertificateWithContext(ctx, uploadCertificateReq)
	c.logger.Debug("sdk request 'ssl.UploadCertificate'", slog.Any("request", uploadCertificateReq), slog.Any("response", uploadCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'ssl.UploadCertificate': %w", err)
	}
	if uploadCertificateResp.Response == nil || lo.FromPtr(uploadCertificateResp.Response.CertificateId) == "" {
		return nil, fmt.Errorf("ssl.UploadCertificate returned no certificate ID")
	}

	// Tencent Cloud deduplicates by the leaf fingerprint. Reuse an existing
	// certificate only when its complete chain also matches the requested one.
	if repeatedID := lo.FromPtr(uploadCertificateResp.Response.RepeatCertId); repeatedID != "" {
		describeReq := tcssl.NewDescribeCertificateDetailRequest()
		describeReq.CertificateId = common.StringPtr(repeatedID)
		describeResp, describeErr := c.sdkClient.DescribeCertificateDetailWithContext(ctx, describeReq)
		if describeErr == nil && describeResp.Response != nil &&
			xcert.EqualCertificateChainsFromPEM(certPEM, lo.FromPtr(describeResp.Response.CertificatePublicKey)) {
			return &UploadResult{CertId: repeatedID}, nil
		}

		// If the chain differs or cannot be read, create a separate cloud record
		// instead of silently deploying the old chain. No extra read permission
		// is required for callers that can already upload certificates.
		uploadCertificateReq.Repeatable = common.BoolPtr(true)
		uploadCertificateResp, err = c.sdkClient.UploadCertificateWithContext(ctx, uploadCertificateReq)
		if err != nil {
			return nil, fmt.Errorf("failed to upload replacement certificate chain: %w", err)
		}
		if uploadCertificateResp.Response == nil || lo.FromPtr(uploadCertificateResp.Response.CertificateId) == "" ||
			lo.FromPtr(uploadCertificateResp.Response.CertificateId) == repeatedID {
			return nil, fmt.Errorf("ssl.UploadCertificate did not create a replacement certificate chain")
		}
	}

	return &UploadResult{
		CertId: *uploadCertificateResp.Response.CertificateId,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*ReplaceResult, error) {
	return nil, core.ErrUnsupported
}

func createSDKClient(secretId, secretKey, endpoint string) (*tcssl.Client, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := tcssl.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
