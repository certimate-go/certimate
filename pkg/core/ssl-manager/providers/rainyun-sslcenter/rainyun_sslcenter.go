package rainyunsslcenter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/certimate-go/certimate/pkg/core"
	rainyunsdk "github.com/certimate-go/certimate/pkg/sdk3rd/rainyun"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type SSLManagerProviderConfig struct {
	// 雨云 API 密钥。
	ApiKey string `json:"ApiKey"`
}

type SSLManagerProvider struct {
	config    *SSLManagerProviderConfig
	logger    *slog.Logger
	sdkClient *rainyunsdk.Client
}

var _ core.SSLManager = (*SSLManagerProvider)(nil)

func NewSSLManagerProvider(config *SSLManagerProviderConfig) (*SSLManagerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl manager provider is nil")
	}

	client, err := createSDKClient(config.ApiKey)
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
	// 遍历证书列表，避免重复上传
	if res, err := m.findCertIfExists(ctx, certPEM); err != nil {
		return nil, err
	} else if res != nil {
		m.logger.Info("ssl certificate already exists")
		return res, nil
	}

	// SSL 证书上传
	// REF: https://apifox.com/apidoc/shared/a4595cc8-44c5-4678-a2a3-eed7738dab03/api-69943046
	sslCenterCreateReq := &rainyunsdk.SslCenterCreateRequest{
		Cert: certPEM,
		Key:  privkeyPEM,
	}
	sslCenterCreateResp, err := m.sdkClient.SslCenterCreate(sslCenterCreateReq)
	m.logger.Debug("sdk request 'sslcenter.Create'", slog.Any("request", sslCenterCreateReq), slog.Any("response", sslCenterCreateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'sslcenter.Create': %w", err)
	}

	// 遍历证书列表，获取刚刚上传证书 ID
	if res, err := m.findCertIfExists(ctx, certPEM); err != nil {
		return nil, err
	} else if res == nil {
		return nil, errors.New("no ssl certificate found, may be upload failed")
	} else {
		return res, nil
	}
}

func (m *SSLManagerProvider) findCertIfExists(ctx context.Context, certPEM string) (*core.SSLManageUploadResult, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 遍历 SSL 证书列表
	// REF: https://apifox.com/apidoc/shared/a4595cc8-44c5-4678-a2a3-eed7738dab03/api-69943046
	// REF: https://apifox.com/apidoc/shared/a4595cc8-44c5-4678-a2a3-eed7738dab03/api-69943048
	sslCenterListPage := int32(1)
	sslCenterListPerPage := int32(100)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		sslCenterListReq := &rainyunsdk.SslCenterListRequest{
			Filters: &rainyunsdk.SslCenterListFilters{
				Domain: &certX509.Subject.CommonName,
			},
			Page:    &sslCenterListPage,
			PerPage: &sslCenterListPerPage,
		}
		sslCenterListResp, err := m.sdkClient.SslCenterList(sslCenterListReq)
		m.logger.Debug("sdk request 'sslcenter.List'", slog.Any("request", sslCenterListReq), slog.Any("response", sslCenterListResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'sslcenter.List': %w", err)
		}

		if sslCenterListResp.Data != nil && sslCenterListResp.Data.Records != nil {
			for _, sslRecord := range sslCenterListResp.Data.Records {
				// 先对比证书的多域名
				if sslRecord.Domain != strings.Join(certX509.DNSNames, ", ") {
					continue
				}

				// 再对比证书的有效期
				if sslRecord.StartDate != certX509.NotBefore.Unix() || sslRecord.ExpireDate != certX509.NotAfter.Unix() {
					continue
				}

				// 最后对比证书内容
				sslCenterGetResp, err := m.sdkClient.SslCenterGet(sslRecord.ID)
				if err != nil {
					return nil, fmt.Errorf("failed to execute sdk request 'sslcenter.Get': %w", err)
				}

				var isSameCert bool
				if sslCenterGetResp.Data != nil {
					if sslCenterGetResp.Data.Cert == certPEM {
						isSameCert = true
					} else {
						oldCertX509, err := xcert.ParseCertificateFromPEM(sslCenterGetResp.Data.Cert)
						if err != nil {
							continue
						}

						isSameCert = xcert.EqualCertificate(certX509, oldCertX509)
					}
				}

				// 如果已存在相同证书，直接返回
				if isSameCert {
					return &core.SSLManageUploadResult{
						CertId: fmt.Sprintf("%d", sslRecord.ID),
					}, nil
				}
			}
		}

		if sslCenterListResp.Data == nil || len(sslCenterListResp.Data.Records) < int(sslCenterListPerPage) {
			break
		} else {
			sslCenterListPage++
		}
	}

	return nil, nil
}

func createSDKClient(apiKey string) (*rainyunsdk.Client, error) {
	return rainyunsdk.NewClient(apiKey)
}
