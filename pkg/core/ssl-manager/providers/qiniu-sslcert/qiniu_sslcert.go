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

	"github.com/certimate-go/certimate/pkg/core"
	qiniusdk "github.com/certimate-go/certimate/pkg/sdk3rd/qiniu"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type SSLManagerProviderConfig struct {
	// 七牛云 AccessKey。
	AccessKey string `json:"accessKey"`
	// 七牛云 SecretKey。
	SecretKey string `json:"secretKey"`
}

type SSLManagerProvider struct {
	config    *SSLManagerProviderConfig
	logger    *slog.Logger
	sdkClient *qiniusdk.SslCertManager
}

var _ core.SSLManager = (*SSLManagerProvider)(nil)

func NewSSLManagerProvider(config *SSLManagerProviderConfig) (*SSLManagerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl manager provider is nil")
	}

	client, err := createSDKClient(config.AccessKey, config.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("could not create sdk client: %w", err)
	}

	return &SSLManagerProvider{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (m *SSLManagerProvider) SetLogger(logger *slog.Logger) {
	if logger == nil {
		m.logger = slog.New(slog.DiscardHandler)
	} else {
		m.logger = logger
	}
}

func (m *SSLManagerProvider) Upload(ctx context.Context, certPEM string, privkeyPEM string) (*core.SSLManageUploadResult, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 生成新证书名（需符合七牛云命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// 遍历查询已有证书，避免重复上传
	getSslCertListMarker := ""
	getSslCertListLimit := int32(200)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		getSslCertListResp, err := m.sdkClient.GetSslCertList(context.TODO(), getSslCertListMarker, getSslCertListLimit)
		m.logger.Debug("sdk request 'sslcert.GetList'", slog.Any("request.marker", getSslCertListMarker), slog.Any("response", getSslCertListResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'sslcert.GetList': %w", err)
		}

		if getSslCertListResp.Certs != nil {
			for _, sslCert := range getSslCertListResp.Certs {
				// 先对比证书通用名称
				if !strings.EqualFold(certX509.Subject.CommonName, sslCert.CommonName) {
					continue
				}

				// 再对比证书多域名
				if !slices.Equal(certX509.DNSNames, sslCert.DnsNames) {
					continue
				}

				// 再对比证书有效期
				if certX509.NotBefore.Unix() != sslCert.NotBefore || certX509.NotAfter.Unix() != sslCert.NotAfter {
					continue
				}

				// 最后对比证书公钥算法
				switch certX509.PublicKeyAlgorithm {
				case x509.RSA:
					if !strings.EqualFold(sslCert.Encrypt, "RSA") {
						continue
					}
				case x509.ECDSA:
					if !strings.EqualFold(sslCert.Encrypt, "ECDSA") {
						continue
					}
				case x509.Ed25519:
					if !strings.EqualFold(sslCert.Encrypt, "ED25519") {
						continue
					}
				default:
					// 未知算法，跳过
					continue
				}

				// 如果以上信息都一致，则视为已存在相同证书，直接返回
				m.logger.Info("ssl certificate already exists")
				return &core.SSLManageUploadResult{
					CertId:   sslCert.CertID,
					CertName: sslCert.Name,
				}, nil
			}
		}

		if len(getSslCertListResp.Certs) < int(getSslCertListLimit) || getSslCertListResp.Marker == "" {
			break
		} else {
			getSslCertListMarker = getSslCertListResp.Marker
		}
	}

	// 上传新证书
	// REF: https://developer.qiniu.com/fusion/8593/interface-related-certificate
	uploadSslCertResp, err := m.sdkClient.UploadSslCert(context.TODO(), certName, certX509.Subject.CommonName, certPEM, privkeyPEM)
	m.logger.Debug("sdk request 'sslcert.Upload'", slog.Any("response", uploadSslCertResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'sslcert.Upload': %w", err)
	}

	return &core.SSLManageUploadResult{
		CertId:   uploadSslCertResp.CertID,
		CertName: certName,
	}, nil
}

func createSDKClient(accessKey, secretKey string) (*qiniusdk.SslCertManager, error) {
	if secretKey == "" {
		return nil, errors.New("invalid qiniu access key")
	}

	if secretKey == "" {
		return nil, errors.New("invalid qiniu secret key")
	}

	credential := auth.New(accessKey, secretKey)
	client := qiniusdk.NewSslCertManager(credential)
	return client, nil
}
