package axisnow

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	axisnowsdk "github.com/certimate-go/certimate/pkg/sdk3rd/axisnow"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type (
	Provider      = core.Certmgr
	UploadResult  = core.CertmgrUploadResult
	ReplaceResult = core.CertmgrReplaceResult
)

type CertmgrConfig struct {
	// AxisNow API 令牌。
	ApiToken string `json:"apiToken"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *axisnowsdk.Client
}

var _ Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.ApiToken)
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
	// 查询证书列表，避免重复上传
	// REF: https://developers.axisnow.io/api-reference/main/client/v1/#/operations/certificate-setting-list-certificates
	listCertificatesPage := 1
	listCertificatesPerPage := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCertificatesReq := &axisnowsdk.ListCertificatesRequest{
			Page:    lo.ToPtr(listCertificatesPage),
			PerPage: lo.ToPtr(listCertificatesPerPage),
		}
		listCertificatesResp, err := c.sdkClient.ListCertificatesWithContext(ctx, listCertificatesReq)
		c.logger.Debug("sdk request 'Certificates.ListCertificates'", slog.Any("request", listCertificatesReq), slog.Any("response", listCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'Certificates.ListCertificates': %w", err)
		} else {
			if listCertificatesResp.Results == nil {
				break
			}

			for _, certItem := range listCertificatesResp.Results {
				if xcert.EqualCertificatesFromPEM(certPEM, certItem.Certificate) {
					// 如果已存在相同证书，直接返回
					c.logger.Info("ssl certificate already exists")
					return &UploadResult{
						CertId:   certItem.UUID,
						CertName: certItem.Name,
					}, nil
				}
			}

			if len(listCertificatesResp.Results) < listCertificatesPerPage ||
				(listCertificatesResp.ResultInfo != nil && listCertificatesPage*listCertificatesPerPage >= listCertificatesResp.ResultInfo.TotalCount) {
				break
			}

			listCertificatesPage++
		}
	}

	// 添加证书
	// REF: https://developers.axisnow.io/api-reference/main/client/v1/#/operations/certificate-setting-add-certificate
	addCertificateReq := &axisnowsdk.AddCertificateRequest{
		Type:        lo.ToPtr("upload"),
		Name:        lo.ToPtr(fmt.Sprintf("certimate-%d", time.Now().UnixMilli())),
		Certificate: lo.ToPtr(certPEM),
		PrivateKey:  lo.ToPtr(privkeyPEM),
	}
	addCertificateResp, err := c.sdkClient.AddCertificateWithContext(ctx, addCertificateReq)
	c.logger.Debug("sdk request 'Certificates.AddCertificate'", slog.Any("request", addCertificateReq), slog.Any("response", addCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'Certificates.AddCertificate': %w", err)
	}

	return &UploadResult{
		CertId:   addCertificateResp.Result.UUID,
		CertName: addCertificateResp.Result.Name,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*ReplaceResult, error) {
	return nil, core.ErrUnsupported
}

func createSDKClient(apiToken string) (*axisnowsdk.Client, error) {
	client, err := axisnowsdk.NewClient(
		axisnowsdk.WithApiToken(apiToken),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
