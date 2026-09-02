package asiaispcdn

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	asiaispcdnsdk "github.com/certimate-go/certimate/pkg/sdk3rd/asiaispcdn"
)

type (
	Provider      = core.Certmgr
	UploadResult  = core.CertmgrUploadResult
	ReplaceResult = core.CertmgrReplaceResult
)

type CertmgrConfig struct {
	// 橙域网络 CDN AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 橙域网络 CDN AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *asiaispcdnsdk.Client
}

var _ Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret)
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
	if upres, upok, err := c.tryGetResultIfCertExists(ctx, certPEM, privkeyPEM); err != nil {
		return nil, err
	} else if upok {
		c.logger.Info("ssl certificate already exists")
		return upres, nil
	}

	// 生成新证书名（需符合橙域网络 CDN 命名规则）
	certName := fmt.Sprintf("certimate_%d", time.Now().UnixMilli())

	// 上传证书
	certificateUploadReq := &asiaispcdnsdk.CertificateUploadRequest{
		PublicKey:  lo.ToPtr(certPEM),
		PrivateKey: lo.ToPtr(privkeyPEM),
		Name:       lo.ToPtr(certName),
	}
	certificateUploadResp, err := c.sdkClient.CertificateUploadWithContext(ctx, certificateUploadReq)
	c.logger.Debug("sdk request 'certificateUpload'", slog.Any("request", certificateUploadReq), slog.Any("response", certificateUploadResp))
	if err != nil {
		if certificateUploadResp != nil && certificateUploadResp.GetCode() == 80003 {
			if upres, upok, err := c.tryGetResultIfCertExists(ctx, certPEM, privkeyPEM); err != nil {
				return nil, err
			} else if !upok {
				return nil, fmt.Errorf("could not find ssl certificate, may be upload failed")
			} else {
				c.logger.Info("ssl certificate already exists")
				return upres, nil
			}
		}

		return nil, fmt.Errorf("failed to execute sdk request 'certificateUpload': %w", err)
	}

	certId, err := certificateUploadResp.Data.Int64()
	if err != nil {
		return nil, fmt.Errorf("received invalid certificate id: %w", err)
	}

	return &UploadResult{
		CertId:   fmt.Sprintf("%d", certId),
		CertName: certName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*ReplaceResult, error) {
	return nil, core.ErrUnsupported
}

func (c *Certmgr) tryGetResultIfCertExists(ctx context.Context, certPEM, privkeyPEM string) (*UploadResult, bool, error) {
	// 查询证书列表
	certificateQueryListResp, err := c.sdkClient.CertificateQueryListWithContext(ctx)
	c.logger.Debug("sdk request 'certificateQueryList'", slog.Any("response", certificateQueryListResp))
	if err != nil {
		return nil, false, fmt.Errorf("failed to execute sdk request 'certificateQueryList': %w", err)
	} else {
		for _, certItem := range certificateQueryListResp.Data {
			certificateQueryResp, err := c.sdkClient.CertificateQueryWithContext(ctx, certItem.CertId)
			c.logger.Debug("sdk request 'certificateQuery'", slog.String("params.certId", certItem.CertId), slog.Any("response", certificateQueryResp))
			if err != nil {
				if certificateQueryResp != nil && certificateQueryResp.GetCode() == 80012 {
					continue
				}

				return nil, false, fmt.Errorf("failed to execute sdk request 'certificateQuery': %w", err)
			}

			if certificateQueryResp.Data.PublicKey == certPEM && certificateQueryResp.Data.PrivateKey == privkeyPEM {
				// 如果已存在相同证书，直接返回
				return &UploadResult{
					CertId:   certItem.CertId,
					CertName: certItem.Name,
				}, true, nil
			}
		}
	}

	return nil, false, nil
}

func createSDKClient(accessKeyId, accessKeySecret string) (*asiaispcdnsdk.Client, error) {
	client, err := asiaispcdnsdk.NewClient(
		asiaispcdnsdk.WithAkSk(accessKeyId, accessKeySecret),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
