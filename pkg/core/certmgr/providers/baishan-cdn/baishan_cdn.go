package baishancdn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	baishansdk "github.com/certimate-go/certimate/pkg/sdk3rd/baishan"
)

type CertmgrConfig struct {
	// 白山云 API Token。
	ApiToken string `json:"apiToken"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *baishansdk.Client
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, errors.New("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.ApiToken)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return &Certmgr{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (d *Certmgr) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}
}

func (d *Certmgr) Upload(ctx context.Context, certPEM, privkeyPEM string) (*certmgr.UploadResult, error) {
	// 生成新证书名（需符合白山云命名规则）
	certName := fmt.Sprintf("certimate_%d", time.Now().UnixMilli())

	// 新增证书
	// REF: https://portal.baishancloud.com/track/document/downloadPdf/1441
	certId := ""
	uploadDomainCertificateReq := &baishansdk.UploadDomainCertificateRequest{
		Name:        lo.ToPtr(certName),
		Certificate: lo.ToPtr(certPEM),
		Key:         lo.ToPtr(privkeyPEM),
	}
	uploadDomainCertificateResp, err := d.sdkClient.UploadDomainCertificateWithContext(ctx, uploadDomainCertificateReq)
	d.logger.Debug("sdk request 'baishan.UploadDomainCertificate'", slog.Any("request", uploadDomainCertificateReq), slog.Any("response", uploadDomainCertificateResp))
	if err != nil {
		if uploadDomainCertificateResp != nil {
			if uploadDomainCertificateResp.GetCode() == 400699 && strings.Contains(uploadDomainCertificateResp.GetMessage(), "this certificate is exists") {
				// 证书已存在，忽略新增证书接口错误
				re := regexp.MustCompile(`\d+`)
				certId = re.FindString(uploadDomainCertificateResp.GetMessage())
			}
		}

		if certId == "" {
			return nil, fmt.Errorf("failed to execute sdk request 'baishan.SetDomainCertificate': %w", err)
		}
	} else {
		certId = uploadDomainCertificateResp.Data.CertId.String()
	}

	return &certmgr.UploadResult{
		CertId:   certId,
		CertName: certName,
	}, nil
}

func (d *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	// 替换证书
	// REF: https://portal.baishancloud.com/track/document/downloadPdf/1441
	uploadDomainCertificateReq := &baishansdk.UploadDomainCertificateRequest{
		CertificateId: lo.ToPtr(certIdOrName),
		Name:          lo.ToPtr(fmt.Sprintf("certimate_%d", time.Now().UnixMilli())),
		Certificate:   lo.ToPtr(certPEM),
		Key:           lo.ToPtr(privkeyPEM),
	}
	uploadDomainCertificateResp, err := d.sdkClient.UploadDomainCertificateWithContext(ctx, uploadDomainCertificateReq)
	d.logger.Debug("sdk request 'baishan.UploadDomainCertificate'", slog.Any("request", uploadDomainCertificateReq), slog.Any("response", uploadDomainCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'baishan.UploadDomainCertificate': %w", err)
	}

	return &certmgr.OperateResult{}, nil
}

func createSDKClient(apiToken string) (*baishansdk.Client, error) {
	return baishansdk.NewClient(apiToken)
}
