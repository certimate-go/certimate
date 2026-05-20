package tencentcloudssl

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	tcssl "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
)

type CertmgrConfig struct {
	// 腾讯云 SecretId。
	SecretId string `json:"secretId"`
	// 腾讯云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 腾讯云接口端点。
	Endpoint string `json:"endpoint,omitempty"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *tcssl.Client
}

var _ certmgr.Provider = (*Certmgr)(nil)

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

func (c *Certmgr) Upload(ctx context.Context, certPEM, privkeyPEM string) (*certmgr.UploadResult, error) {
	// 上传新证书
	// REF: https://cloud.tencent.com/document/api/400/41665
	uploadCertificateReq := tcssl.NewUploadCertificateRequest()
	uploadCertificateReq.CertificatePublicKey = common.StringPtr(certPEM)
	uploadCertificateReq.CertificatePrivateKey = common.StringPtr(privkeyPEM)
	uploadCertificateReq.Repeatable = common.BoolPtr(false)
	uploadCertificateResp, err := c.sdkClient.UploadCertificateWithContext(ctx, uploadCertificateReq)
	c.logger.Debug("sdk request 'ssl.UploadCertificate'", slog.Any("request", uploadCertificateReq), slog.Any("response", uploadCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'ssl.UploadCertificate': %w", err)
	}

	return &certmgr.UploadResult{
		CertId: *uploadCertificateResp.Response.CertificateId,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	return nil, certmgr.ErrUnsupported
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
