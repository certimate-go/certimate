package digitaloceancertificate

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	digitaloceansdk "github.com/certimate-go/certimate/pkg/sdk3rd/digitalocean"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type (
	Provider      = core.Certmgr
	UploadResult  = core.CertmgrUploadResult
	ReplaceResult = core.CertmgrReplaceResult
)

type CertmgrConfig struct {
	// DigitalOcean AccessToken。
	AccessToken string `json:"accessToken"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *digitaloceansdk.Client
}

var _ Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.AccessToken)
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

func (c *Certmgr) Upload(ctx context.Context, certPEM, privkeyPEM string) (*UploadResult, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 提取服务器证书和中间证书
	serverCertPEM, issuerCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract certs: %w", err)
	}

	// 查询证书列表，避免重复上传
	// REF: https://docs.digitalocean.com/reference/api/reference/certificates/#certificates_list
	listCertificatesPage := 1
	listCertificatesPerPage := 200
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCertificatesReq := &digitaloceansdk.ListCertificatesRequest{
			Page:    lo.ToPtr(listCertificatesPage),
			PerPage: lo.ToPtr(listCertificatesPerPage),
		}
		listCertificatesResp, err := c.sdkClient.ListCertificatesWithContext(ctx, listCertificatesReq)
		c.logger.Debug("sdk request 'ListCertificates'", slog.Any("request", listCertificatesReq), slog.Any("response", listCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'ListCertificates': %w", err)
		}

		fingerprintSha1 := sha1.Sum(certX509.Raw)
		fingerprintSha1Hex := hex.EncodeToString(fingerprintSha1[:])
		for _, certItem := range listCertificatesResp.Certificates {
			// 对比证书扩展名称
			if !slices.Equal(certX509.DNSNames, certItem.DNSNames) {
				continue
			}

			// 对比证书有效期
			newCertNotAfter := certX509.NotAfter
			oldCertNotAfter, _ := time.Parse("2006-01-02T15:04:05Z", certItem.NotAfter)
			if !newCertNotAfter.Equal(oldCertNotAfter) {
				continue
			}

			// 对比证书指纹
			if !strings.EqualFold(fingerprintSha1Hex, certItem.SHA1Fingerprint) {
				continue
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			c.logger.Info("ssl certificate already exists")
			return &UploadResult{
				CertId:   certItem.ID,
				CertName: certItem.Name,
			}, nil
		}

		if len(listCertificatesResp.Certificates) < listCertificatesPerPage {
			break
		}

		listCertificatesPage++
	}

	// 创建新证书
	// REF: https://docs.digitalocean.com/reference/api/reference/certificates/#certificates_create
	createCertificateReq := &digitaloceansdk.CreateCertificateRequest{
		Name:             lo.ToPtr(fmt.Sprintf("certimate-%d", time.Now().UnixMilli())),
		Type:             lo.ToPtr("custom"),
		CertificateChain: lo.ToPtr(issuerCertPEM),
		LeafCertificate:  lo.ToPtr(serverCertPEM),
		PrivateKey:       lo.ToPtr(privkeyPEM),
	}
	createCertificateResp, err := c.sdkClient.CreateCertificateWithContext(ctx, createCertificateReq)
	c.logger.Debug("sdk request 'CreateCertificate'", slog.Any("request", createCertificateReq), slog.Any("response", createCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'CreateCertificate': %w", err)
	}

	return &UploadResult{
		CertId:   createCertificateResp.Certificate.ID,
		CertName: createCertificateResp.Certificate.Name,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*ReplaceResult, error) {
	return nil, core.ErrUnsupported
}

func createSDKClient(accessToken string) (*digitaloceansdk.Client, error) {
	client, err := digitaloceansdk.NewClient(
		digitaloceansdk.WithAccessToken(accessToken),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
