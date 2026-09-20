package autossl

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/certimate-go/certimate/pkg/core"
	autosslsdk "github.com/certimate-go/certimate/pkg/sdk3rd/autossl"
)

type (
	Provider    = core.Certsource
	FetchResult = core.CertsourceFetchResult
)

type CertsourceConfig struct {
	// AutoSSL AccessKey。
	AccessKey string `json:"accessKey"`
	// AutoSSL SecretKey。
	SecretKey string `json:"secretKey"`
}

type Certsource struct {
	config    *CertsourceConfig
	logger    *slog.Logger
	sdkClient *autosslsdk.Client
}

var _ Provider = (*Certsource)(nil)

func NewCertsource(config *CertsourceConfig) (*Certsource, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the certsource provider is nil")
	}

	client, err := createSDKClient(config.AccessKey, config.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return &Certsource{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (c *Certsource) SetLogger(logger *slog.Logger) {
	if logger == nil {
		c.logger = slog.New(slog.DiscardHandler)
	} else {
		c.logger = logger
	}
}

func (c *Certsource) Fetch(ctx context.Context, domain string, certId string) (*FetchResult, error) {
	if certId == "" && domain == "" {
		return nil, fmt.Errorf("the configuration item 'Domain' and 'CertId' are both empty")
	}

	var matchedCertId string
	if certId != "" {
		// 指定了证书 ID，直接采用
		matchedCertId = certId
	} else {
		// 按域名匹配证书列表，取仍未过期且截止时间最晚的一条
		items, err := c.sdkClient.ListAllCertificatesWithContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'cert.getCertList': %w", err)
		}

		now := time.Now()
		var matched *autosslsdk.CertificateInfo
		for _, item := range items {
			if item.Url != domain {
				continue
			}

			if !item.CertEndTime.IsZero() && !item.CertEndTime.After(now) {
				continue
			}

			if matched == nil || item.CertEndTime.After(matched.CertEndTime) {
				matched = item
			}
		}

		if matched == nil {
			return nil, fmt.Errorf("no valid certificate found for domain '%s'", domain)
		}

		matchedCertId = matched.CertId
		c.logger.Info(fmt.Sprintf("matched certificate #%s for domain '%s'", matchedCertId, domain))
	}

	// 下载证书内容
	// REF: https://www.auto-ssl.cn/docs/OpenAPI.html#downloadcert
	downloadResp, err := c.sdkClient.DownloadCertWithContext(ctx, &autosslsdk.DownloadCertRequest{
		Id: matchedCertId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cert.downloadCert': %w", err)
	}

	keypair := downloadResp.GetKeypair()
	if keypair == nil || keypair.CertPEM == "" || keypair.PrivkeyPEM == "" {
		return nil, fmt.Errorf("the downloaded certificate content of certificate #%s is empty", matchedCertId)
	}

	// 规范化换行符
	certPEM := strings.ReplaceAll(strings.TrimSpace(keypair.CertPEM), "\r\n", "\n")
	privkeyPEM := strings.ReplaceAll(strings.TrimSpace(keypair.PrivkeyPEM), "\r\n", "\n")

	return &FetchResult{
		CertPEM:    certPEM,
		PrivkeyPEM: privkeyPEM,
		ExtendedData: map[string]any{
			"certId": matchedCertId,
		},
	}, nil
}

func createSDKClient(accessKey, secretKey string) (*autosslsdk.Client, error) {
	client, err := autosslsdk.NewClient(
		autosslsdk.WithAccessKey(accessKey),
		autosslsdk.WithSecretKey(secretKey),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
