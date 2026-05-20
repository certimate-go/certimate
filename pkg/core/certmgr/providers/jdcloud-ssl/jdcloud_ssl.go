package jdcloudssl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	jdcore "github.com/jdcloud-api/jdcloud-sdk-go/core"
	jdsslapis "github.com/jdcloud-api/jdcloud-sdk-go/services/ssl/apis"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	jdssl "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/jdcloud-api/jdcloud-sdk-go/services/ssl/client"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type CertmgrConfig struct {
	// 京东云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 京东云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *jdssl.SslClient
}

var _ certmgr.Provider = (*Certmgr)(nil)

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

func (c *Certmgr) Upload(ctx context.Context, certPEM, privkeyPEM string) (*certmgr.UploadResult, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 格式化私钥内容，以便后续计算私钥摘要
	privkeyPEM = strings.TrimSpace(privkeyPEM)
	privkeyPEM = strings.ReplaceAll(privkeyPEM, "\r", "")
	privkeyPEM = strings.ReplaceAll(privkeyPEM, "\n", "\r\n")
	privkeyPEM = privkeyPEM + "\r\n"

	// 查看证书列表
	// REF: https://docs.jdcloud.com/cn/ssl-certificate/api/describecerts
	describeCertsPageNumber := 1
	describeCertsPageSize := 10
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		describeCertsReq := jdsslapis.NewDescribeCertsRequestWithoutParam()
		describeCertsReq.SetDomainName(certX509.Subject.CommonName)
		describeCertsReq.SetPageNumber(describeCertsPageNumber)
		describeCertsReq.SetPageSize(describeCertsPageSize)
		describeCertsResp, err := c.sdkClient.DescribeCerts(describeCertsReq)
		c.logger.Debug("sdk request 'ssl.DescribeCerts'", slog.Any("request", describeCertsReq), slog.Any("response", describeCertsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'ssl.DescribeCerts': %w", err)
		}

		for _, certItem := range describeCertsResp.Result.CertListDetails {
			// 对比证书通用名称
			if !strings.EqualFold(certX509.Subject.CommonName, certItem.CommonName) {
				continue
			}

			// 对比证书多域名
			if !strings.EqualFold(strings.Join(certX509.DNSNames, ","), strings.Join(certItem.DnsNames, ",")) {
				continue
			}

			// 对比证书有效期
			oldCertNotBefore, _ := time.Parse(time.RFC3339, certItem.StartTime)
			oldCertNotAfter, _ := time.Parse(time.RFC3339, certItem.EndTime)
			if !certX509.NotBefore.Equal(oldCertNotBefore) || !certX509.NotAfter.Equal(oldCertNotAfter) {
				continue
			}

			// 对比私钥 SHA-256 摘要
			newKeyDigest := sha256.Sum256([]byte(privkeyPEM))
			newKeyDigestHex := hex.EncodeToString(newKeyDigest[:])
			if !strings.EqualFold(newKeyDigestHex, certItem.Digest) {
				continue
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			c.logger.Info("ssl certificate already exists")
			return &certmgr.UploadResult{
				CertId:   certItem.CertId,
				CertName: certItem.CertName,
			}, nil
		}

		if len(describeCertsResp.Result.CertListDetails) < int(describeCertsPageSize) {
			break
		} else {
			describeCertsPageNumber++
		}
	}

	// 生成新证书名（需符合京东云命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// 上传证书
	// REF: https://docs.jdcloud.com/cn/ssl-certificate/api/uploadcert
	uploadCertReq := jdsslapis.NewUploadCertRequestWithoutParam()
	uploadCertReq.SetCertName(certName)
	uploadCertReq.SetCertFile(certPEM)
	uploadCertReq.SetKeyFile(privkeyPEM)
	uploadCertResp, err := c.sdkClient.UploadCert(uploadCertReq)
	c.logger.Debug("sdk request 'ssl.UploadCertificate'", slog.Any("request", uploadCertReq), slog.Any("response", uploadCertResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'ssl.UploadCertificate': %w", err)
	}

	return &certmgr.UploadResult{
		CertId:   uploadCertResp.Result.CertId,
		CertName: certName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.ReplaceResult, error) {
	return nil, certmgr.ErrUnsupported
}

func createSDKClient(accessKeyId, accessKeySecret string) (*jdssl.SslClient, error) {
	clientCredentials := jdcore.NewCredentials(accessKeyId, accessKeySecret)
	client := jdssl.NewSslClient(clientCredentials)
	client.DisableLogger()
	return client, nil
}
