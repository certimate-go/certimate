package ksyunslb

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/ksyun-kcm"
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
	// 金山云地域。
	Region string `json:"region"`
}

type Certmgr struct {
	config     *CertmgrConfig
	logger     *slog.Logger
	sdkClient  *ksyunkcmsdk.Client
	sdkCertmgr core.Certmgr
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

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		SecretAccessKey: config.SecretAccessKey,
		ProjectId:       config.ProjectId,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create certmgr: %w", err)
	}

	return &Certmgr{
		config:     config,
		logger:     slog.Default(),
		sdkClient:  client,
		sdkCertmgr: pcertmgr,
	}, nil
}

func (c *Certmgr) SetLogger(logger *slog.Logger) {
	if logger == nil {
		c.logger = slog.New(slog.DiscardHandler)
	} else {
		c.logger = logger
	}

	c.sdkCertmgr.SetLogger(logger)
}

func (c *Certmgr) Upload(ctx context.Context, certPEM, privkeyPEM string) (*UploadResult, error) {
	// 描述证书，避免重复上传
	describeCertificatesPage := 1
	describeCertificatesPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		describeCertificatesReq := &ksyunkcmsdk.DescribeCertificatesRequest{
			Region:   lo.ToPtr(c.config.Region),
			Page:     lo.ToPtr(int32(describeCertificatesPage)),
			PageSize: lo.ToPtr(int32(describeCertificatesPageSize)),
		}
		describeCertificatesResp, err := c.sdkClient.DescribeCertificatesWithContext(ctx, describeCertificatesReq)
		c.logger.Debug("sdk request 'kcm.DescribeCertificates'", slog.Any("request", describeCertificatesReq), slog.Any("response", describeCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'kcm.DescribeCertificates': %w", err)
		}

		if describeCertificatesResp.CertificateSet == nil {
			break
		}

		for _, certItem := range describeCertificatesResp.CertificateSet {
			// 如果已存在相同证书，直接返回
			if xcert.EqualCertificatesFromPEM(certPEM, certItem.PublicKey) {
				c.logger.Info("ssl certificate already exists")
				return &UploadResult{
					CertId:   certItem.CertificateId,
					CertName: certItem.CertificateName,
				}, nil
			}
		}

		if len(describeCertificatesResp.CertificateSet) < describeCertificatesPageSize {
			break
		}

		describeCertificatesPage++
	}

	// 托管证书到 KCM
	upres, err := c.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	}

	// 创建证书
	// REF: https://apiexplorer.ksyun.com/#/api/96/CreateCertificate/2016-03-04/1013
	createCertificateReq := &ksyunkcmsdk.CreateCertificateRequest{
		Region:           lo.ToPtr(c.config.Region),
		CertificateName:  lo.ToPtr(fmt.Sprintf("certimate-%d", time.Now().UnixMilli())),
		Description:      lo.ToPtr("upload from certimate"),
		Source:           lo.ToPtr("kcm"),
		SSLCertificateId: lo.ToPtr(upres.CertId),
	}
	createCertificateResp, err := c.sdkClient.CreateCertificateWithContext(ctx, createCertificateReq)
	c.logger.Debug("sdk request 'kcm.CreateCertificate'", slog.Any("request", createCertificateReq), slog.Any("response", createCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'kcm.CreateCertificate': %w", err)
	}

	return &UploadResult{
		CertId:   createCertificateResp.Certificate.CertificateId,
		CertName: createCertificateResp.Certificate.CertificateName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*ReplaceResult, error) {
	// 托管证书到 KCM
	upres, err := c.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	}

	// 更新证书
	// REF: https://apiexplorer.ksyun.com/#/api/96/ModifyCertificate/2016-03-04/1013
	modifyCertificateReq := &ksyunkcmsdk.ModifyCertificateRequest{
		Region:           lo.ToPtr(c.config.Region),
		Description:      lo.ToPtr("upload from certimate"),
		SSLCertificateId: lo.ToPtr(upres.CertId),
	}
	modifyCertificateResp, err := c.sdkClient.ModifyCertificateWithContext(ctx, modifyCertificateReq)
	c.logger.Debug("sdk request 'kcm.ModifyCertificate'", slog.Any("request", modifyCertificateReq), slog.Any("response", modifyCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'kcm.ModifyCertificate': %w", err)
	}

	return &ReplaceResult{}, nil
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
