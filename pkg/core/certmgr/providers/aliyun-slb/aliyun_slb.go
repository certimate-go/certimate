package aliyunslb

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	alislb "github.com/alibabacloud-go/slb-20140515/v4/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	"github.com/certimate-go/certimate/pkg/core/certmgr/providers/aliyun-slb/internal"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type CertmgrConfig struct {
	// 阿里云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 阿里云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 阿里云资源组 ID。
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
	// 阿里云地域。
	Region string `json:"region"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *internal.SlbClient
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, errors.New("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret, config.Region)
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
	// REF: https://help.aliyun.com/zh/slb/classic-load-balancer/developer-reference/api-slb-2014-05-15-describeservercertificates
	describeServerCertificatesReq := &alislb.DescribeServerCertificatesRequest{
		ResourceGroupId: lo.EmptyableToPtr(c.config.ResourceGroupId),
		RegionId:        tea.String(c.config.Region),
	}
	describeServerCertificatesResp, err := c.sdkClient.DescribeServerCertificatesWithContext(ctx, describeServerCertificatesReq, &dara.RuntimeOptions{})
	c.logger.Debug("sdk request 'slb.DescribeServerCertificates'", slog.Any("request", describeServerCertificatesReq), slog.Any("response", describeServerCertificatesResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'slb.DescribeServerCertificates': %w", err)
	}

	if describeServerCertificatesResp.Body.ServerCertificates != nil && describeServerCertificatesResp.Body.ServerCertificates.ServerCertificate != nil {
		fingerprintSha256 := sha256.Sum256(certX509.Raw)
		fingerprintSha256Hex := hex.EncodeToString(fingerprintSha256[:])
		fingerprintSha1 := sha1.Sum(certX509.Raw)
		fingerprintSha1Hex := hex.EncodeToString(fingerprintSha1[:])
		for _, certItem := range describeServerCertificatesResp.Body.ServerCertificates.ServerCertificate {
			if tea.Int32Value(certItem.IsAliCloudCertificate) != 0 {
				continue
			}

			// 对比证书通用名称
			if !strings.EqualFold(certX509.Subject.CommonName, tea.StringValue(certItem.CommonName)) {
				continue
			}

			// 对比证书 SHA-1 或 SHA-256 摘要
			oldFingerprint := strings.ReplaceAll(tea.StringValue(certItem.Fingerprint), ":", "")
			if !strings.EqualFold(fingerprintSha256Hex, oldFingerprint) && !strings.EqualFold(fingerprintSha1Hex, oldFingerprint) {
				continue
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			c.logger.Info("ssl certificate already exists")
			return &certmgr.UploadResult{
				CertId:   *certItem.ServerCertificateId,
				CertName: *certItem.ServerCertificateName,
			}, nil
		}
	}

	// 生成新证书名（需符合阿里云命名规则）
	certName := fmt.Sprintf("certimate_%d", time.Now().UnixMilli())

	// 去除证书和私钥内容中的空白行，以符合阿里云 API 要求
	// REF: https://github.com/certimate-go/certimate/issues/326
	re := regexp.MustCompile(`(?m)^\s*$\n?`)
	certPEM = strings.TrimSpace(re.ReplaceAllString(certPEM, ""))
	privkeyPEM = strings.TrimSpace(re.ReplaceAllString(privkeyPEM, ""))

	// 上传新证书
	// REF: https://help.aliyun.com/zh/slb/classic-load-balancer/developer-reference/api-slb-2014-05-15-uploadservercertificate
	uploadServerCertificateReq := &alislb.UploadServerCertificateRequest{
		ResourceGroupId:       lo.EmptyableToPtr(c.config.ResourceGroupId),
		RegionId:              tea.String(c.config.Region),
		ServerCertificateName: tea.String(certName),
		ServerCertificate:     tea.String(certPEM),
		PrivateKey:            tea.String(privkeyPEM),
	}
	uploadServerCertificateResp, err := c.sdkClient.UploadServerCertificateWithContext(ctx, uploadServerCertificateReq, &dara.RuntimeOptions{})
	c.logger.Debug("sdk request 'slb.UploadServerCertificate'", slog.Any("request", uploadServerCertificateReq), slog.Any("response", uploadServerCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'slb.UploadServerCertificate': %w", err)
	}

	return &certmgr.UploadResult{
		CertId:   *uploadServerCertificateResp.Body.ServerCertificateId,
		CertName: certName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	return nil, certmgr.ErrUnsupported
}

func createSDKClient(accessKeyId, accessKeySecret, region string) (*internal.SlbClient, error) {
	// 接入点一览 https://api.aliyun.com/product/Slb
	var endpoint string
	switch region {
	case "",
		"cn-hangzhou",
		"cn-hangzhou-finance",
		"cn-shanghai-finance-1",
		"cn-shenzhen-finance-1":
		endpoint = "slb.aliyuncs.com"
	default:
		endpoint = fmt.Sprintf("slb.%s.aliyuncs.com", region)
	}

	config := &aliopen.Config{
		Endpoint:        tea.String(endpoint),
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
	}

	client, err := internal.NewSlbClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
