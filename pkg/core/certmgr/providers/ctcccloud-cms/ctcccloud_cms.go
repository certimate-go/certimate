package ctcccloudcms

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	ctyuncms "github.com/certimate-go/certimate/pkg/sdk3rd/ctyun/cms"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type CertmgrConfig struct {
	// 天翼云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 天翼云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *ctyuncms.Client
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, errors.New("the configuration of the certmgr provider is nil")
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

func (c *Certmgr) Upload(ctx context.Context, certPEM, privkeyPEM string) (*certmgr.UploadResult, error) {
	// 避免重复上传
	if upres, upok, err := c.tryGetResultIfCertExists(ctx, certPEM); err != nil {
		return nil, err
	} else if upok {
		c.logger.Info("ssl certificate already exists")
		return upres, nil
	}

	// 提取服务器证书和中间证书
	serverCertPEM, intermediaCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract certs: %w", err)
	}

	// 生成新证书名（需符合天翼云命名规则）
	certName := fmt.Sprintf("cm%d", time.Now().Unix())

	// 上传证书
	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=152&api=17243&data=204&isNormal=1&vid=283
	uploadCertificateReq := &ctyuncms.UploadCertificateRequest{
		Name:               lo.ToPtr(certName),
		Certificate:        lo.ToPtr(serverCertPEM),
		CertificateChain:   lo.ToPtr(intermediaCertPEM),
		PrivateKey:         lo.ToPtr(privkeyPEM),
		EncryptionStandard: lo.ToPtr("INTERNATIONAL"),
	}
	uploadCertificateResp, err := c.sdkClient.UploadCertificate(uploadCertificateReq)
	c.logger.Debug("sdk request 'cms.UploadCertificate'", slog.Any("request", uploadCertificateReq), slog.Any("response", uploadCertificateResp))
	if err != nil {
		if uploadCertificateResp != nil && uploadCertificateResp.GetError() == "CCMS_100000067" {
			if upres, upok, err := c.tryGetResultIfCertExists(ctx, certPEM); err != nil {
				return nil, err
			} else if !upok {
				return nil, errors.New("ctyun cms: no certificate found")
			} else {
				c.logger.Info("ssl certificate already exists")
				return upres, nil
			}
		}

		return nil, fmt.Errorf("failed to execute sdk request 'cms.UploadCertificate': %w", err)
	}

	// 获取刚刚上传证书 ID
	if upres, upok, err := c.tryGetResultIfCertExists(ctx, certPEM); err != nil {
		return nil, err
	} else if !upok {
		return nil, fmt.Errorf("could not find ssl certificate, may be upload failed")
	} else {
		return upres, nil
	}
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	return nil, certmgr.ErrUnsupported
}

func (c *Certmgr) tryGetResultIfCertExists(ctx context.Context, certPEM string) (*certmgr.UploadResult, bool, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, false, err
	}

	// 查询用户证书列表
	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=152&api=17233&data=204&isNormal=1&vid=283
	getCertificateListPageNum := 1
	getCertificateListPageSize := 10
	for {
		select {
		case <-ctx.Done():
			return nil, false, ctx.Err()
		default:
		}

		getCertificateListReq := &ctyuncms.GetCertificateListRequest{
			PageNum:  lo.ToPtr(int32(getCertificateListPageNum)),
			PageSize: lo.ToPtr(int32(getCertificateListPageSize)),
			Keyword:  lo.ToPtr(certX509.Subject.CommonName),
			Origin:   lo.ToPtr("UPLOAD"),
		}
		getCertificateListResp, err := c.sdkClient.GetCertificateList(getCertificateListReq)
		c.logger.Debug("sdk request 'cms.GetCertificateList'", slog.Any("request", getCertificateListReq), slog.Any("response", getCertificateListResp))
		if err != nil {
			return nil, false, fmt.Errorf("failed to execute sdk request 'cms.GetCertificateList': %w", err)
		}

		if getCertificateListResp.ReturnObj == nil {
			break
		}

		for _, certItem := range getCertificateListResp.ReturnObj.List {
			// 对比证书名称
			if !strings.EqualFold(strings.Join(certX509.DNSNames, ","), certItem.DomainName) {
				continue
			}

			// 对比证书有效期
			oldCertNotBefore, _ := time.Parse("2006-01-02T15:04:05Z", certItem.IssueTime)
			oldCertNotAfter, _ := time.Parse("2006-01-02T15:04:05Z", certItem.ExpireTime)
			if !certX509.NotBefore.Equal(oldCertNotBefore) {
				continue
			} else if !certX509.NotAfter.Equal(oldCertNotAfter) {
				continue
			}

			// 对比证书指纹
			fingerprint := sha1.Sum(certX509.Raw)
			fingerprintHex := hex.EncodeToString(fingerprint[:])
			if !strings.EqualFold(fingerprintHex, certItem.Fingerprint) {
				continue
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			c.logger.Info("ssl certificate already exists")
			return &certmgr.UploadResult{
				CertId:   certItem.Id,
				CertName: certItem.Name,
			}, true, nil
		}

		if len(getCertificateListResp.ReturnObj.List) < getCertificateListPageSize {
			break
		}

		getCertificateListPageNum++
	}

	return nil, false, nil
}

func createSDKClient(accessKeyId, secretAccessKey string) (*ctyuncms.Client, error) {
	return ctyuncms.NewClient(accessKeyId, secretAccessKey)
}
