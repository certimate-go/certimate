package ksyunkcm

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	ksyunkcmsdk "github.com/certimate-go/certimate/pkg/sdk3rd/ksyun/kcm"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type (
	Provider      = core.Certmgr
	UploadResult  = core.CertmgrUploadResult
	ReplaceResult = core.CertmgrReplaceResult
)

type CertmgrConfig struct {
	// 金山云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 金山云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 金山云项目 ID。
	ProjectId int64 `json:"projectId,omitempty"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *ksyunkcmsdk.Client
}

var _ Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey)
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
	// 避免重复上传
	if upres, upok, err := c.tryGetResultIfCertExists(ctx, certPEM); err != nil {
		return nil, err
	} else if upok {
		c.logger.Info("ssl certificate already exists")
		return upres, nil
	}

	// 上传证书
	uploadCertificateReq := &ksyunkcmsdk.UploadCertificateRequest{
		ProjectId: lo.ToPtr(c.config.ProjectId),
		CertName:  lo.ToPtr(fmt.Sprintf("certimate-%d", time.Now().UnixMilli())),
		CertFile:  lo.ToPtr(certPEM),
		CertKey:   lo.ToPtr(privkeyPEM),
	}
	uploadCertificateResp, err := c.sdkClient.UploadCertificateWithContext(ctx, uploadCertificateReq)
	c.logger.Debug("sdk request 'kcm.UploadCertificate'", slog.Any("request", uploadCertificateReq), slog.Any("response", uploadCertificateResp))
	if err != nil {
		if uploadCertificateResp != nil &&
			uploadCertificateResp.Error != nil && uploadCertificateResp.Error.Message == "重复的证书文件" {
			if upres, upok, err := c.tryGetResultIfCertExists(ctx, certPEM); err != nil {
				return nil, err
			} else if !upok {
				return nil, fmt.Errorf("could not find ssl certificate, may be upload failed")
			} else {
				c.logger.Info("ssl certificate already exists")
				return upres, nil
			}
		}

		return nil, fmt.Errorf("failed to execute sdk request 'kcm.UploadCertificate': %w", err)
	}

	return &UploadResult{
		CertId:   uploadCertificateResp.Ret.CertId,
		CertName: uploadCertificateResp.Ret.CertName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*ReplaceResult, error) {
	return nil, core.ErrUnsupported
}

func (c *Certmgr) tryGetResultIfCertExists(ctx context.Context, certPEM string) (*UploadResult, bool, error) {
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, false, err
	}

	listUserCertificatesPage := 1
	listUserCertificatesPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, false, ctx.Err()
		default:
		}

		listUserCertificatesReq := &ksyunkcmsdk.ListUserCertificatesRequest{
			Page:     lo.ToPtr(int32(listUserCertificatesPage)),
			PageSize: lo.ToPtr(int32(listUserCertificatesPageSize)),
		}
		listUserCertificatesResp, err := c.sdkClient.ListUserCertificatesWithContext(ctx, listUserCertificatesReq)
		c.logger.Debug("sdk request 'kcm.ListUserCertificates'", slog.Any("request", listUserCertificatesReq), slog.Any("response", listUserCertificatesResp))
		if err != nil {
			return nil, false, fmt.Errorf("failed to execute sdk request 'kcm.ListUserCertificates': %w", err)
		}

		if listUserCertificatesResp.Ret == nil || listUserCertificatesResp.Ret.Certs == nil {
			break
		}

		fingerprintSha1 := sha1.Sum(certX509.Raw)
		fingerprintSha1Hex := hex.EncodeToString(fingerprintSha1[:])
		for _, certItem := range listUserCertificatesResp.Ret.Certs {
			// 对比证书备用名称
			if !strings.EqualFold(strings.Join(certX509.DNSNames, ","), strings.Join(certItem.Domains, ",")) {
				continue
			}

			// 对比证书颁发者
			if certX509.Issuer.CommonName != certItem.CA {
				continue
			}

			// 对比证书指纹
			if !strings.EqualFold(fingerprintSha1Hex, certItem.FingerPrint) {
				continue
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			return &UploadResult{
				CertId:   certItem.CertId,
				CertName: certItem.CertName,
			}, true, nil
		}

		if len(listUserCertificatesResp.Ret.Certs) < listUserCertificatesPageSize {
			break
		}

		listUserCertificatesPage++
	}

	return nil, false, nil
}

func createSDKClient(accessKeyId, secretAccessKey string) (*ksyunkcmsdk.Client, error) {
	client, err := ksyunkcmsdk.NewClient(
		ksyunkcmsdk.WithAkSk(accessKeyId, secretAccessKey),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
