package aliyunslb

import (
	"context"
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
	"github.com/alibabacloud-go/tea/tea"

	"github.com/certimate-go/certimate/pkg/core"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xtypes "github.com/certimate-go/certimate/pkg/utils/types"
)

type SSLManagerProviderConfig struct {
	// 阿里云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 阿里云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 阿里云资源组 ID。
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
	// 阿里云地域。
	Region string `json:"region"`
}

type SSLManagerProvider struct {
	config    *SSLManagerProviderConfig
	logger    *slog.Logger
	sdkClient *alislb.Client
}

var _ core.SSLManager = (*SSLManagerProvider)(nil)

func NewSSLManagerProvider(config *SSLManagerProviderConfig) (*SSLManagerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl manager provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret, config.Region)
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

	// 查询证书列表，避免重复上传
	// REF: https://help.aliyun.com/zh/slb/classic-load-balancer/developer-reference/api-slb-2014-05-15-describeservercertificates
	describeServerCertificatesReq := &alislb.DescribeServerCertificatesRequest{
		ResourceGroupId: xtypes.ToPtrOrZeroNil(m.config.ResourceGroupId),
		RegionId:        tea.String(m.config.Region),
	}
	describeServerCertificatesResp, err := m.sdkClient.DescribeServerCertificates(describeServerCertificatesReq)
	m.logger.Debug("sdk request 'slb.DescribeServerCertificates'", slog.Any("request", describeServerCertificatesReq), slog.Any("response", describeServerCertificatesResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'slb.DescribeServerCertificates': %w", err)
	}

	if describeServerCertificatesResp.Body.ServerCertificates != nil && describeServerCertificatesResp.Body.ServerCertificates.ServerCertificate != nil {
		fingerprint := sha256.Sum256(certX509.Raw)
		fingerprintHex := hex.EncodeToString(fingerprint[:])
		for _, serverCert := range describeServerCertificatesResp.Body.ServerCertificates.ServerCertificate {
			isSameCert := *serverCert.IsAliCloudCertificate == 0 &&
				strings.EqualFold(fingerprintHex, strings.ReplaceAll(*serverCert.Fingerprint, ":", "")) &&
				strings.EqualFold(certX509.Subject.CommonName, *serverCert.CommonName)
			// 如果已存在相同证书，直接返回
			if isSameCert {
				m.logger.Info("ssl certificate already exists")
				return &core.SSLManageUploadResult{
					CertId:   *serverCert.ServerCertificateId,
					CertName: *serverCert.ServerCertificateName,
				}, nil
			}
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
		ResourceGroupId:       xtypes.ToPtrOrZeroNil(m.config.ResourceGroupId),
		RegionId:              tea.String(m.config.Region),
		ServerCertificateName: tea.String(certName),
		ServerCertificate:     tea.String(certPEM),
		PrivateKey:            tea.String(privkeyPEM),
	}
	uploadServerCertificateResp, err := m.sdkClient.UploadServerCertificate(uploadServerCertificateReq)
	m.logger.Debug("sdk request 'slb.UploadServerCertificate'", slog.Any("request", uploadServerCertificateReq), slog.Any("response", uploadServerCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'slb.UploadServerCertificate': %w", err)
	}

	return &core.SSLManageUploadResult{
		CertId:   *uploadServerCertificateResp.Body.ServerCertificateId,
		CertName: certName,
	}, nil
}

func createSDKClient(accessKeyId, accessKeySecret, region string) (*alislb.Client, error) {
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

	client, err := alislb.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
