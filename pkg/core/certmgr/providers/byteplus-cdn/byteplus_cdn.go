package bytepluscdn

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	bpcdn "github.com/byteplus-sdk/byteplus-sdk-golang/service/cdn"

	"github.com/certimate-go/certimate/pkg/core"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type (
	Provider      = core.Certmgr
	UploadResult  = core.CertmgrUploadResult
	ReplaceResult = core.CertmgrReplaceResult
)

type CertmgrConfig struct {
	// BytePlus AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// BytePlus SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *bpcdn.CDN
}

var _ Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the certmgr provider is nil")
	}

	client := bpcdn.NewInstance()
	client.Client.SetAccessKey(config.AccessKeyId)
	client.Client.SetSecretKey(config.SecretAccessKey)

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

	// 查询证书列表，避免重复上传
	// REF: https://docs.byteplus.com/en/docs/byteplus-cdn/reference-listcertinfo
	listCertInfoPageNum := 1
	listCertInfoPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCertInfoReq := &bpcdn.ListCertInfoRequest{
			PageNum:  bpcdn.GetInt64Ptr(int64(listCertInfoPageNum)),
			PageSize: bpcdn.GetInt64Ptr(int64(listCertInfoPageSize)),
			Source:   bpcdn.GetStrPtr("cert_center"),
		}
		listCertInfoResp, err := c.sdkClient.ListCertInfo(listCertInfoReq)
		c.logger.Debug("sdk request 'cdn.ListCertInfo'", slog.Any("request", listCertInfoReq), slog.Any("response", listCertInfoResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'cdn.ListCertInfo': %w", err)
		}

		for _, certItem := range listCertInfoResp.Result.CertInfo {
			// 对比证书 SHA-1 摘要
			fingerprintSha1 := sha1.Sum(certX509.Raw)
			if !strings.EqualFold(hex.EncodeToString(fingerprintSha1[:]), certItem.CertFingerprint.Sha1) {
				continue
			}

			// 对比证书 SHA-256 摘要
			fingerprintSha256 := sha256.Sum256(certX509.Raw)
			if !strings.EqualFold(hex.EncodeToString(fingerprintSha256[:]), certItem.CertFingerprint.Sha256) {
				continue
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			c.logger.Info("ssl certificate already exists")
			return &UploadResult{
				CertId:   certItem.CertId,
				CertName: certItem.Desc,
			}, nil
		}

		if len(listCertInfoResp.Result.CertInfo) < listCertInfoPageSize {
			break
		}

		listCertInfoPageNum++
	}

	// 生成新证书名（需符合 BytePlus 命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// 上传新证书
	// REF: https://docs.byteplus.com/en/docs/byteplus-cdn/reference-addcertificate
	addCertificateReq := &bpcdn.AddCertificateRequest{
		Certificate: certPEM,
		PrivateKey:  privkeyPEM,
		Source:      bpcdn.GetStrPtr("cert_center"),
		Desc:        bpcdn.GetStrPtr(certName),
	}
	addCertificateResp, err := c.sdkClient.AddCertificate(addCertificateReq)
	c.logger.Debug("sdk request 'cdn.AddCertificate'", slog.Any("request", addCertificateReq), slog.Any("response", addCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cdn.AddCertificate': %w", err)
	}

	return &UploadResult{
		CertId:   addCertificateResp.Result.CertId,
		CertName: certName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*ReplaceResult, error) {
	return nil, core.ErrUnsupported
}
