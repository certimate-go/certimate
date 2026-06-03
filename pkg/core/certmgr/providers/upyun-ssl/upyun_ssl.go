package upyunssl

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/pkg/core"
	upyunsdk "github.com/certimate-go/certimate/pkg/sdk3rd/upyun/console"
)

type (
	Provider      = core.Certmgr
	UploadResult  = core.CertmgrUploadResult
	ReplaceResult = core.CertmgrReplaceResult
)

type CertmgrConfig struct {
	// 又拍云账号用户名。
	Username string `json:"username"`
	// 又拍云账号密码。
	Password string `json:"password"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *upyunsdk.Client
}

var _ Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.Username, config.Password)
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
	// 上传证书
	uploadHttpsCertificateReq := &upyunsdk.UploadHttpsCertificateRequest{
		Certificate: certPEM,
		PrivateKey:  privkeyPEM,
	}
	uploadHttpsCertificateResp, err := c.sdkClient.UploadHttpsCertificateWithContext(ctx, uploadHttpsCertificateReq)
	c.logger.Debug("sdk request 'console.UploadHttpsCertificate'", slog.Any("response", uploadHttpsCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'console.UploadHttpsCertificate': %w", err)
	}

	return &UploadResult{
		CertId: uploadHttpsCertificateResp.Data.Result.CertificateId,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*ReplaceResult, error) {
	return nil, core.ErrUnsupported
}

func createSDKClient(username, password string) (*upyunsdk.Client, error) {
	return upyunsdk.NewClient(username, password)
}
