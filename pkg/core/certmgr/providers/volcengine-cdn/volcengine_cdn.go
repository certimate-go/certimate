package volcenginecdn

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	vecdn "github.com/volcengine/volcengine-go-sdk/service/cdn"
	ve "github.com/volcengine/volcengine-go-sdk/volcengine"
	vesession "github.com/volcengine/volcengine-go-sdk/volcengine/session"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	"github.com/certimate-go/certimate/pkg/core/certmgr/providers/volcengine-cdn/internal"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type CertmgrConfig struct {
	// 火山引擎 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 火山引擎 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *internal.CdnClient
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, errors.New("the configuration of the certmgr provider is nil")
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

func (c *Certmgr) Upload(ctx context.Context, certPEM, privkeyPEM string) (*certmgr.UploadResult, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 查询证书列表，避免重复上传
	// REF: https://www.volcengine.com/docs/6454/125709
	listCertInfoPageNum := 1
	listCertInfoPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCertInfoReq := &vecdn.ListCertInfoInput{
			Source:   ve.String("volc_cert_center"),
			PageNum:  ve.Int32(int32(listCertInfoPageNum)),
			PageSize: ve.Int32(int32(listCertInfoPageSize)),
		}
		listCertInfoResp, err := c.sdkClient.ListCertInfo(listCertInfoReq)
		c.logger.Debug("sdk request 'cdn.ListCertInfo'", slog.Any("request", listCertInfoReq), slog.Any("response", listCertInfoResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'cdn.ListCertInfo': %w", err)
		}

		for _, certItem := range listCertInfoResp.CertInfo {
			// 对比证书 SHA-1 摘要
			fingerprintSha1 := sha1.Sum(certX509.Raw)
			if !strings.EqualFold(hex.EncodeToString(fingerprintSha1[:]), ve.StringValue(certItem.CertFingerprint.Sha1)) {
				continue
			}

			// 对比证书 SHA-256 摘要
			fingerprintSha256 := sha256.Sum256(certX509.Raw)
			if !strings.EqualFold(hex.EncodeToString(fingerprintSha256[:]), ve.StringValue(certItem.CertFingerprint.Sha256)) {
				continue
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			c.logger.Info("ssl certificate already exists")
			return &certmgr.UploadResult{
				CertId:   ve.StringValue(certItem.CertId),
				CertName: ve.StringValue(certItem.Desc),
			}, nil
		}

		if len(listCertInfoResp.CertInfo) < listCertInfoPageSize {
			break
		}

		listCertInfoPageNum++
	}

	// 生成新证书名（需符合火山引擎命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// 上传新证书
	// REF: https://www.volcengine.com/docs/6454/1245763
	addCertificateReq := &vecdn.AddCertificateInput{
		Source:      ve.String("volc_cert_center"),
		Certificate: ve.String(certPEM),
		PrivateKey:  ve.String(privkeyPEM),
		Desc:        ve.String(certName),
	}
	addCertificateResp, err := c.sdkClient.AddCertificate(addCertificateReq)
	c.logger.Debug("sdk request 'cdn.AddCertificate'", slog.Any("request", addCertificateResp), slog.Any("response", addCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cdn.AddCertificate': %w", err)
	}

	return &certmgr.UploadResult{
		CertId:   ve.StringValue(addCertificateResp.CertId),
		CertName: certName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	return nil, certmgr.ErrUnsupported
}

func createSDKClient(accessKeyId, accessKeySecret string) (*internal.CdnClient, error) {
	config := ve.NewConfig().
		WithAkSk(accessKeyId, accessKeySecret).
		WithRegion("cn-north-1")

	session, err := vesession.NewSession(config)
	if err != nil {
		return nil, err
	}

	client := internal.NewCdnClient(session)
	return client, nil
}
