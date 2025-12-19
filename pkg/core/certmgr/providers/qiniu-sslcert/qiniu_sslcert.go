package qiniusslcert

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/qiniu/go-sdk/v7/auth"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	qiniusdk "github.com/certimate-go/certimate/pkg/sdk3rd/qiniu"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type CertmgrConfig struct {
	// 七牛云 AccessKey。
	AccessKey string `json:"accessKey"`
	// 七牛云 SecretKey。
	SecretKey string `json:"secretKey"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *qiniusdk.SslCertManager
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, errors.New("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.AccessKey, config.SecretKey)
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

	// 生成新证书名（需符合七牛云命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// 查询已有证书，避免重复上传
	getSslCertListMarker := ""
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		getSslCertListResp, err := c.sdkClient.GetSslCertList(ctx, getSslCertListMarker, 200)
		c.logger.Debug("sdk request 'sslcert.GetList'", slog.Any("request.marker", getSslCertListMarker), slog.Any("response", getSslCertListResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'sslcert.GetList': %w", err)
		}

		for _, sslItem := range getSslCertListResp.Certs {
			// 对比证书通用名称
			if !strings.EqualFold(certX509.Subject.CommonName, sslItem.CommonName) {
				continue
			}

			// 对比证书多域名
			if !slices.Equal(certX509.DNSNames, sslItem.DnsNames) {
				continue
			}

			// 对比证书有效期
			if certX509.NotBefore.Unix() != sslItem.NotBefore || certX509.NotAfter.Unix() != sslItem.NotAfter {
				continue
			}

			// 对比证书公钥算法
			switch certX509.PublicKeyAlgorithm {
			case x509.RSA:
				if !strings.EqualFold(sslItem.Encrypt, "RSA") {
					continue
				}
			case x509.ECDSA:
				if !strings.EqualFold(sslItem.Encrypt, "ECDSA") {
					continue
				}
			case x509.Ed25519:
				if !strings.EqualFold(sslItem.Encrypt, "Ed25519") {
					continue
				}
			default:
				// 未知算法，跳过
				continue
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			c.logger.Info("ssl certificate already exists")
			return &certmgr.UploadResult{
				CertId:   sslItem.CertID,
				CertName: sslItem.Name,
			}, nil
		}

		if len(getSslCertListResp.Certs) == 0 || getSslCertListResp.Marker == "" {
			break
		}

		getSslCertListMarker = getSslCertListResp.Marker
	}

	// 上传新证书
	// REF: https://developer.qiniu.com/fusion/8593/interface-related-certificate
	uploadSslCertResp, err := c.sdkClient.UploadSslCert(ctx, certName, certX509.Subject.CommonName, certPEM, privkeyPEM)
	c.logger.Debug("sdk request 'sslcert.Upload'", slog.Any("response", uploadSslCertResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'sslcert.Upload': %w", err)
	}

	return &certmgr.UploadResult{
		CertId:   uploadSslCertResp.CertID,
		CertName: certName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	return nil, certmgr.ErrUnsupported
}

func createSDKClient(accessKey, secretKey string) (*qiniusdk.SslCertManager, error) {
	if secretKey == "" {
		return nil, errors.New("qiniu: invalid access key")
	}
	if secretKey == "" {
		return nil, errors.New("qiniu: invalid secret key")
	}

	credential := auth.New(accessKey, secretKey)
	client := qiniusdk.NewSslCertManager(credential)
	return client, nil
}
