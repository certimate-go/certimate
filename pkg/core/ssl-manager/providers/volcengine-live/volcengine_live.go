package volcenginelive

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	velive "github.com/volcengine/volc-sdk-golang/service/live/v20230101"
	ve "github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/certimate-go/certimate/pkg/core"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type SSLManagerProviderConfig struct {
	// 火山引擎 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 火山引擎 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
}

type SSLManagerProvider struct {
	config    *SSLManagerProviderConfig
	logger    *slog.Logger
	sdkClient *velive.Live
}

var _ core.SSLManager = (*SSLManagerProvider)(nil)

func NewSSLManagerProvider(config *SSLManagerProviderConfig) (*SSLManagerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl manager provider is nil")
	}

	client := velive.NewInstance()
	client.SetAccessKey(config.AccessKeyId)
	client.SetSecretKey(config.AccessKeySecret)

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
	// REF: https://www.volcengine.com/docs/6469/1186278#%E6%9F%A5%E8%AF%A2%E8%AF%81%E4%B9%A6%E5%88%97%E8%A1%A8
	listCertReq := &velive.ListCertV2Body{}
	listCertResp, err := m.sdkClient.ListCertV2(ctx, listCertReq)
	m.logger.Debug("sdk request 'live.ListCertV2'", slog.Any("request", listCertReq), slog.Any("response", listCertResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'live.ListCertV2': %w", err)
	}
	if listCertResp.Result.CertList != nil {
		for _, certInfo := range listCertResp.Result.CertList {
			// 查询证书详细信息
			// REF: https://www.volcengine.com/docs/6469/1186278#%E6%9F%A5%E7%9C%8B%E8%AF%81%E4%B9%A6%E8%AF%A6%E6%83%85
			describeCertDetailSecretReq := &velive.DescribeCertDetailSecretV2Body{
				ChainID: ve.String(certInfo.ChainID),
			}
			describeCertDetailSecretResp, err := m.sdkClient.DescribeCertDetailSecretV2(ctx, describeCertDetailSecretReq)
			m.logger.Debug("sdk request 'live.DescribeCertDetailSecretV2'", slog.Any("request", describeCertDetailSecretReq), slog.Any("response", describeCertDetailSecretResp))
			if err != nil {
				continue
			}

			var isSameCert bool
			certificate := strings.Join(describeCertDetailSecretResp.Result.SSL.Chain, "\n\n")
			if certificate == certPEM {
				isSameCert = true
			} else {
				oldCertX509, err := xcert.ParseCertificateFromPEM(certificate)
				if err != nil {
					continue
				}

				isSameCert = xcert.EqualCertificate(certX509, oldCertX509)
			}

			// 如果已存在相同证书，直接返回
			if isSameCert {
				m.logger.Info("ssl certificate already exists")
				return &core.SSLManageUploadResult{
					CertId:   certInfo.ChainID,
					CertName: certInfo.CertName,
				}, nil
			}
		}
	}

	// 生成新证书名（需符合火山引擎命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// 上传新证书
	// REF: https://www.volcengine.com/docs/6469/1186278#%E6%B7%BB%E5%8A%A0%E8%AF%81%E4%B9%A6
	createCertReq := &velive.CreateCertBody{
		CertName:    ve.String(certName),
		UseWay:      "https",
		ProjectName: ve.String("default"),
		Rsa: velive.CreateCertBodyRsa{
			Prikey: privkeyPEM,
			Pubkey: certPEM,
		},
	}
	createCertResp, err := m.sdkClient.CreateCert(ctx, createCertReq)
	m.logger.Debug("sdk request 'live.CreateCert'", slog.Any("request", createCertReq), slog.Any("response", createCertResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'live.CreateCert': %w", err)
	}

	return &core.SSLManageUploadResult{
		CertId:   *createCertResp.Result.ChainID,
		CertName: certName,
	}, nil
}
