package wangsucertificate

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/certimate-go/certimate/pkg/core"
	wangsusdk "github.com/certimate-go/certimate/pkg/sdk3rd/wangsu/certificate"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xtypes "github.com/certimate-go/certimate/pkg/utils/types"
)

type SSLManagerProviderConfig struct {
	// 网宿云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 网宿云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
}

type SSLManagerProvider struct {
	config    *SSLManagerProviderConfig
	logger    *slog.Logger
	sdkClient *wangsusdk.Client
}

var _ core.SSLManager = (*SSLManagerProvider)(nil)

func NewSSLManagerProvider(config *SSLManagerProviderConfig) (*SSLManagerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl manager provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret)
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
	// REF: https://www.wangsu.com/document/api-doc/22675?productCode=certificatemanagement
	listCertificatesResp, err := m.sdkClient.ListCertificates()
	m.logger.Debug("sdk request 'certificatemanagement.ListCertificates'", slog.Any("response", listCertificatesResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'certificatemanagement.ListCertificates': %w", err)
	}

	if listCertificatesResp.Certificates != nil {
		for _, certRecord := range listCertificatesResp.Certificates {
			// 对比证书序列号
			if !strings.EqualFold(certX509.SerialNumber.Text(16), certRecord.Serial) {
				continue
			}

			// 再对比证书有效期
			cstzone := time.FixedZone("CST", 8*60*60)
			oldCertNotBefore, _ := time.ParseInLocation(time.DateTime, certRecord.ValidityFrom, cstzone)
			oldCertNotAfter, _ := time.ParseInLocation(time.DateTime, certRecord.ValidityTo, cstzone)
			if !certX509.NotBefore.Equal(oldCertNotBefore) || !certX509.NotAfter.Equal(oldCertNotAfter) {
				continue
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			m.logger.Info("ssl certificate already exists")
			return &core.SSLManageUploadResult{
				CertId:   certRecord.CertificateId,
				CertName: certRecord.Name,
			}, nil
		}
	}

	// 生成新证书名（需符合网宿云命名规则）
	certName := fmt.Sprintf("certimate_%d", time.Now().UnixMilli())

	// 新增证书
	// REF: https://www.wangsu.com/document/api-doc/25199?productCode=certificatemanagement
	createCertificateReq := &wangsusdk.CreateCertificateRequest{
		Name:        xtypes.ToPtr(certName),
		Certificate: xtypes.ToPtr(certPEM),
		PrivateKey:  xtypes.ToPtr(privkeyPEM),
		Comment:     xtypes.ToPtr("upload from certimate"),
	}
	createCertificateResp, err := m.sdkClient.CreateCertificate(createCertificateReq)
	m.logger.Debug("sdk request 'certificatemanagement.CreateCertificate'", slog.Any("request", createCertificateReq), slog.Any("response", createCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'certificatemanagement.CreateCertificate': %w", err)
	}

	// 网宿云证书 URL 中包含证书 ID
	// 格式：
	//    https://open.chinanetcenter.com/api/certificate/100001
	wangsuCertIdMatches := regexp.MustCompile(`/certificate/([0-9]+)`).FindStringSubmatch(createCertificateResp.CertificateLocation)
	if len(wangsuCertIdMatches) == 0 {
		return nil, fmt.Errorf("received empty certificate id")
	}

	return &core.SSLManageUploadResult{
		CertId:   wangsuCertIdMatches[1],
		CertName: certName,
	}, nil
}

func createSDKClient(accessKeyId, accessKeySecret string) (*wangsusdk.Client, error) {
	return wangsusdk.NewClient(accessKeyId, accessKeySecret)
}
