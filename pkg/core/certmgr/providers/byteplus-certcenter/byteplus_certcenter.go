package bytepluscertcenter

import (
	"context"
	"fmt"
	"log/slog"

	bp "github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus"
	bpsesion "github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/session"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	bpcertificateservice "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/byteplus-sdk/byteplus-go-sdk-v2/service/certificateservice"
)

type CertmgrConfig struct {
	// BytePlus AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// BytePlus SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// BytePlus 项目名称。
	ProjectName string `json:"projectName,omitempty"`
	// BytePlus 地域。
	Region string `json:"region"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *bpcertificateservice.CERTIFICATESERVICE
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the certmgr provider is nil")
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
	// 上传证书
	// REF: https://docs.byteplus.com/en/docs/byteplus-certificate-center/reference-uploadcertificate
	uploadCertificateReq := &bpcertificateservice.UploadCertificateInput{
		ProjectName: lo.EmptyableToPtr(c.config.ProjectName),
		CertificateInfo: &bpcertificateservice.CertificateInfoForUploadCertificateInput{
			CertificateChain: bp.String(certPEM),
			PrivateKey:       bp.String(privkeyPEM),
		},
		Repeatable: bp.Bool(false),
	}
	uploadCertificateResp, err := c.sdkClient.UploadCertificateWithContext(ctx, uploadCertificateReq)
	c.logger.Debug("sdk request 'certificateservice.UploadCertificate'", slog.Any("request", uploadCertificateReq), slog.Any("response", uploadCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'certificateservice.UploadCertificate': %w", err)
	}

	var sslId string
	if uploadCertificateResp.InstanceId != nil && *uploadCertificateResp.InstanceId != "" {
		sslId = *uploadCertificateResp.InstanceId
	}
	if uploadCertificateResp.RepeatId != nil && *uploadCertificateResp.RepeatId != "" {
		sslId = *uploadCertificateResp.RepeatId
	}

	if sslId == "" {
		return nil, fmt.Errorf("received empty certificate id, both `InstanceId` and `RepeatId` are empty")
	}

	return &certmgr.UploadResult{
		CertId: sslId,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.ReplaceResult, error) {
	return nil, certmgr.ErrUnsupported
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*bpcertificateservice.CERTIFICATESERVICE, error) {
	if region == "" {
		region = "ap-singapore-1" // 证书中心默认区域：新加坡
	}

	config := bp.NewConfig().
		WithAkSk(accessKeyId, secretAccessKey).
		WithRegion(region)

	session, err := bpsesion.NewSession(config)
	if err != nil {
		return nil, err
	}

	client := bpcertificateservice.New(session, config)
	return client, nil
}
