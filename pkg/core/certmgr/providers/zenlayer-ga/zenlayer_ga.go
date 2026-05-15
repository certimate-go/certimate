package zenlayerga

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	zcommon "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zga "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zga20230706"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	"github.com/certimate-go/certimate/pkg/core/certmgr/providers/zenlayer-ga/internal"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type CertmgrConfig struct {
	// Zenlayer AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// Zenlayer AccessKeyPassword。
	AccessKeyPassword string `json:"accessKeyPassword"`
	// Zenlayer 资源组 ID。
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *internal.ZgaClient
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, errors.New("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeyPassword)
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
	// REF: https://docs.console.zenlayer.com/api-reference/cn/networking/zga/certificate/describecertificates
	describeCertificatesPageNum := 1
	describeCertificatesPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		describeCertificatesReq := zga.NewDescribeCertificatesRequest()
		describeCertificatesReq.ResourceGroupId = c.config.ResourceGroupId
		describeCertificatesReq.PageNum = describeCertificatesPageNum
		describeCertificatesReq.PageSize = describeCertificatesPageSize
		describeCertificatesResp, err := c.sdkClient.DescribeCertificates(describeCertificatesReq)
		c.logger.Debug("sdk request 'zga.DescribeCertificates'", slog.Any("request", describeCertificatesReq), slog.Any("response", describeCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'zga.DescribeCertificates': %w", err)
		}

		for _, certItem := range describeCertificatesResp.Response.DataSet {
			// 对比证书通用名称
			if !strings.EqualFold(certX509.Subject.CommonName, certItem.Common) {
				continue
			}

			// 对比证书扩展名称
			if !slices.Equal(certX509.DNSNames, certItem.Sans) {
				continue
			}

			// 对比证书有效期
			oldCertNotBefore, _ := time.Parse("2006-01-02T15:04:05Z", certItem.StartTime)
			oldCertNotAfter, _ := time.Parse("2006-01-02T15:04:05Z", certItem.EndTime)
			if !certX509.NotBefore.Equal(oldCertNotBefore) {
				continue
			} else if !certX509.NotAfter.Equal(oldCertNotAfter) {
				continue
			}

			// 对比证书指纹
			//
			// 注意，虽然文档中描述为 MD5 摘要，但示例给出的是 SHA-1 摘要，因此这里都尝试对比一下
			fingerprintMd5 := md5.Sum(certX509.Raw)
			fingerprintMd5Hex := hex.EncodeToString(fingerprintMd5[:])
			fingerprintSha1 := sha1.Sum(certX509.Raw)
			fingerprintSha1Hex := hex.EncodeToString(fingerprintSha1[:])
			if !strings.EqualFold(fingerprintMd5Hex, certItem.Fingerprint) && !strings.EqualFold(fingerprintSha1Hex, certItem.Fingerprint) {
				continue
			}

			// 对比证书公钥算法
			switch certX509.PublicKeyAlgorithm {
			case x509.RSA:
				if !strings.EqualFold(certItem.Algorithm, "RSA") {
					continue
				}
			case x509.ECDSA:
				if !strings.EqualFold(certItem.Algorithm, "EC") && !strings.EqualFold(certItem.Algorithm, "ECDSA") {
					continue
				}
			default:
				// 未知算法，跳过
				continue
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			c.logger.Info("ssl certificate already exists")
			return &certmgr.UploadResult{
				CertId:   certItem.CertificateId,
				CertName: certItem.CertificateLabel,
			}, nil
		}

		if len(describeCertificatesResp.Response.DataSet) < describeCertificatesPageSize {
			break
		}

		describeCertificatesPageNum++
	}

	// 生成新证书名（需符合 Zenlayer 命名规则）
	certName := fmt.Sprintf("certimate_%d", time.Now().UnixMilli())

	// 创建证书
	// REF: https://docs.console.zenlayer.com/api-reference/cn/networking/zga/certificate/createcertificate
	createCertificateReq := zga.NewCreateCertificateRequest()
	createCertificateReq.ResourceGroupId = c.config.ResourceGroupId
	createCertificateReq.CertificateLabel = certName
	createCertificateReq.CertificateContent = certPEM
	createCertificateReq.CertificateKey = privkeyPEM
	createCertificateResp, err := c.sdkClient.CreateCertificate(createCertificateReq)
	c.logger.Debug("sdk request 'zga.CreateCertificate'", slog.Any("request", createCertificateReq), slog.Any("response", createCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'zga.CreateCertificate': %w", err)
	}

	return &certmgr.UploadResult{
		CertId:   createCertificateResp.Response.CertificateId,
		CertName: certName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	// 更新证书
	// REF: https://docs.console.zenlayer.com/api-reference/cn/networking/zga/certificate/modifycertificate
	modifyCertificateReq := zga.NewModifyCertificateRequest()
	modifyCertificateReq.CertificateId = certIdOrName
	modifyCertificateReq.CertificateContent = certPEM
	modifyCertificateReq.CertificateKey = privkeyPEM
	modifyCertificateResp, err := c.sdkClient.ModifyCertificate(modifyCertificateReq)
	c.logger.Debug("sdk request 'zga.ModifyCertificate'", slog.Any("request", modifyCertificateReq), slog.Any("response", modifyCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'zga.ModifyCertificate': %w", err)
	}

	return &certmgr.OperateResult{}, nil
}

func createSDKClient(accessKeyId, accessKeyPassword string) (*internal.ZgaClient, error) {
	config := zcommon.NewConfig()

	client, err := internal.NewZgaClient(config, accessKeyId, accessKeyPassword)
	if err != nil {
		return nil, err
	}

	return client, nil
}
