package baishancdn

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/certimate-go/certimate/pkg/core"
	bssdk "github.com/certimate-go/certimate/pkg/sdk3rd/baishan"
	xtypes "github.com/certimate-go/certimate/pkg/utils/types"
)

type SSLDeployerProviderConfig struct {
	// 白山云 API Token。
	ApiToken string `json:"apiToken"`
	// 加速域名（支持泛域名）。
	Domain string `json:"domain"`
	// 证书 ID。
	// 选填。零值时表示新建证书；否则表示更新证书。
	CertificateId string `json:"certificateId,omitempty"`
}

type SSLDeployerProvider struct {
	config    *SSLDeployerProviderConfig
	logger    *slog.Logger
	sdkClient *bssdk.Client
}

var _ core.SSLDeployer = (*SSLDeployerProvider)(nil)

func NewSSLDeployerProvider(config *SSLDeployerProviderConfig) (*SSLDeployerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl deployer provider is nil")
	}

	client, err := createSDKClient(config.ApiToken)
	if err != nil {
		return nil, fmt.Errorf("could not create sdk client: %w", err)
	}

	return &SSLDeployerProvider{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (d *SSLDeployerProvider) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}
}

func (d *SSLDeployerProvider) Deploy(ctx context.Context, certPEM string, privkeyPEM string) (*core.SSLDeployResult, error) {
	if d.config.Domain == "" {
		return nil, errors.New("config `domain` is required")
	}

	// 如果原证书 ID 为空，则新增证书；否则替换证书。
	if d.config.CertificateId == "" {
		// 新增证书
		// REF: https://portal.baishancloud.com/track/document/downloadPdf/1441
		certificateId := ""
		setDomainCertificateReq := &bssdk.SetDomainCertificateRequest{
			Name:        xtypes.ToPtr(fmt.Sprintf("certimate_%d", time.Now().UnixMilli())),
			Certificate: xtypes.ToPtr(certPEM),
			Key:         xtypes.ToPtr(privkeyPEM),
		}
		setDomainCertificateResp, err := d.sdkClient.SetDomainCertificate(setDomainCertificateReq)
		d.logger.Debug("sdk request 'baishan.SetDomainCertificate'", slog.Any("request", setDomainCertificateReq), slog.Any("response", setDomainCertificateResp))
		if err != nil {
			if setDomainCertificateResp != nil {
				if setDomainCertificateResp.GetCode() == 400699 && strings.Contains(setDomainCertificateResp.GetMessage(), "this certificate is exists") {
					// 证书已存在，忽略新增证书接口错误
					re := regexp.MustCompile(`\d+`)
					certificateId = re.FindString(setDomainCertificateResp.GetMessage())
				}
			}

			if certificateId == "" {
				return nil, fmt.Errorf("failed to execute sdk request 'baishan.SetDomainCertificate': %w", err)
			}
		} else {
			certificateId = setDomainCertificateResp.Data.CertId.String()
		}

		// 查询域名配置
		// REF: https://portal.baishancloud.com/track/document/api/1/1065
		getDomainConfigReq := &bssdk.GetDomainConfigRequest{
			Domains: xtypes.ToPtr(d.config.Domain),
			Config:  xtypes.ToPtr([]string{"https"}),
		}
		getDomainConfigResp, err := d.sdkClient.GetDomainConfig(getDomainConfigReq)
		d.logger.Debug("sdk request 'baishan.GetDomainConfig'", slog.Any("request", getDomainConfigReq), slog.Any("response", getDomainConfigResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'baishan.GetDomainConfig': %w", err)
		} else if len(getDomainConfigResp.Data) == 0 {
			return nil, fmt.Errorf("domain %s not found", d.config.Domain)
		}

		// 设置域名配置
		// REF: https://portal.baishancloud.com/track/document/api/1/1045
		setDomainConfigReq := &bssdk.SetDomainConfigRequest{
			Domains: xtypes.ToPtr(d.config.Domain),
			Config: &bssdk.DomainConfig{
				Https: &bssdk.DomainConfigHttps{
					CertId:      json.Number(certificateId),
					ForceHttps:  getDomainConfigResp.Data[0].Config.Https.ForceHttps,
					EnableHttp2: getDomainConfigResp.Data[0].Config.Https.EnableHttp2,
					EnableOcsp:  getDomainConfigResp.Data[0].Config.Https.EnableOcsp,
				},
			},
		}
		setDomainConfigResp, err := d.sdkClient.SetDomainConfig(setDomainConfigReq)
		d.logger.Debug("sdk request 'baishan.SetDomainConfig'", slog.Any("request", setDomainConfigReq), slog.Any("response", setDomainConfigResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'baishan.SetDomainConfig': %w", err)
		}
	} else {
		// 替换证书
		// REF: https://portal.baishancloud.com/track/document/downloadPdf/1441
		setDomainCertificateReq := &bssdk.SetDomainCertificateRequest{
			CertificateId: &d.config.CertificateId,
			Name:          xtypes.ToPtr(fmt.Sprintf("certimate_%d", time.Now().UnixMilli())),
			Certificate:   xtypes.ToPtr(certPEM),
			Key:           xtypes.ToPtr(privkeyPEM),
		}
		setDomainCertificateResp, err := d.sdkClient.SetDomainCertificate(setDomainCertificateReq)
		d.logger.Debug("sdk request 'baishan.SetDomainCertificate'", slog.Any("request", setDomainCertificateReq), slog.Any("response", setDomainCertificateResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'baishan.SetDomainCertificate': %w", err)
		}
	}

	return &core.SSLDeployResult{}, nil
}

func createSDKClient(apiToken string) (*bssdk.Client, error) {
	return bssdk.NewClient(apiToken)
}
